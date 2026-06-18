package donation

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"time"

	"social-be/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// defaultImportCurrency is applied to every imported donation. The Excel sheet
// does not carry a currency column and the amounts are Indonesian Rupiah.
const defaultImportCurrency = "IDR"

// Rollback safety errors.
var (
	ErrBatchNotFound        = errors.New("import batch not found")
	ErrBatchAlreadyRolledBk = errors.New("import batch already rolled back")
)

// ImportService handles bulk creation of donations from an uploaded Excel file.
//
// Processing model:
//   - Invalid rows are reported and skipped; they never abort the import.
//   - Valid rows are inserted in batches inside a single transaction together
//     with the audit log, so data and log stay consistent.
//   - Lookups are bulk-loaded once (no per-row queries).
//   - Concurrent imports of the same file are serialised via an advisory lock,
//     and duplicates are re-checked under that lock.
type ImportService struct {
	Repo Repository
}

func NewImportService(repo Repository) *ImportService {
	return &ImportService{Repo: repo}
}

// Import parses the Excel stream, validates every row against bulk-loaded
// master data, and (unless DryRun) persists the valid donations plus an audit
// log. It never stops on the first error.
func (s *ImportService) Import(ctx context.Context, file io.Reader, opts ImportOptions) (*DonationImportResult, error) {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"event":    "donation_import",
		"filename": opts.Filename,
		"user_id":  derefInt(opts.UploadedBy),
		"dry_run":  opts.DryRun,
	})

	sheet, err := parseDonationSheet(file)
	if err != nil {
		log.WithError(err).Warn("import rejected: invalid file/template")
		return nil, err // template/structure error — surfaced to the uploader
	}

	result := &DonationImportResult{
		DryRun:         opts.DryRun,
		Filename:       opts.Filename,
		KomisariatID:   sheet.Meta.KomisariatID,
		KomisariatName: sheet.Meta.KomisariatName,
		Period:         sheet.Meta.Period,
		TotalRows:      len(sheet.Rows),
		Errors:         []DonationImportError{},
		Warnings:       []DonationImportError{},
	}
	log.WithField("total_rows", result.TotalRows).Info("import started")

	// --- Bulk-load master data (constant number of queries) ---------------
	donorIDs := distinctDonaturIDs(sheet.Rows)
	donors, err := s.Repo.FindDonaturRefsByIDs(ctx, donorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load donors: %w", err)
	}
	categories, err := s.Repo.LoadCategoryIndex(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load donation categories: %w", err)
	}

	donorPKs := make([]int, 0, len(donors))
	for _, d := range donors {
		donorPKs = append(donorPKs, d.ID)
	}

	// --- Validate rows in memory (no duplicate decision yet) --------------
	candidates := make([]ImportCandidate, 0, len(sheet.Rows))
	errorRows := make([]ImportError, 0)
	for _, row := range sheet.Rows {
		cand, rowErr := s.classifyRow(row, donors, categories)
		if rowErr != nil {
			result.Errors = append(result.Errors, *rowErr)
			result.FailedRows++
			errorRows = append(errorRows, ImportError{
				RowNumber:    row.Row,
				DonaturID:    row.DonaturID,
				CategoryName: row.CategoryName,
				ErrorMessage: rowErr.Message,
			})
			continue
		}
		candidates = append(candidates, *cand)
	}
	result.Status = importStatus(opts.DryRun, result.FailedRows)

	// --- Dry run: detect duplicates in-memory, write nothing --------------
	if opts.DryRun {
		existing, err := s.Repo.LoadExistingDedupKeys(ctx, donorPKs)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing donations: %w", err)
		}
		seen := make(map[string]struct{}, len(candidates))
		for _, c := range candidates {
			if _, dup := existing[c.DedupKey]; dup {
				result.Warnings = append(result.Warnings, DonationImportError{Row: c.Row, Message: "Donation already exists"})
				result.SkippedRows++
				continue
			}
			if _, dup := seen[c.DedupKey]; dup {
				result.Warnings = append(result.Warnings, DonationImportError{Row: c.Row, Message: "Donation already exists"})
				result.SkippedRows++
				continue
			}
			seen[c.DedupKey] = struct{}{}
			result.SuccessRows++
		}
		log.WithFields(logrus.Fields{
			"success_rows": result.SuccessRows,
			"failed_rows":  result.FailedRows,
			"skipped_rows": result.SkippedRows,
		}).Info("import finished (dry run)")
		return result, nil
	}

	// --- Real run: persist race-free under an advisory lock ---------------
	batchID := uuid.NewString()
	result.BatchID = batchID
	logRecord := s.buildLog(batchID, opts, sheet.Meta, result)
	lockKey := fileLockKey(sheet.Meta, opts.Filename)

	for i := range candidates {
		candidates[i].Donation.ImportBatchID = batchID
	}
	for i := range errorRows {
		errorRows[i].BatchID = batchID
	}

	outcome, err := s.Repo.PersistImportLocked(ctx, lockKey, donorPKs, candidates, errorRows, logRecord)
	if err != nil {
		// Insert/commit failed: no donations were committed. Record a failed
		// audit log (best effort) so the attempt is traceable.
		result.SuccessRows = 0
		result.Status = StatusFailed
		result.BatchID = ""
		failedLog := s.buildLog(batchID, opts, sheet.Meta, result)
		_ = s.Repo.CreateImportLog(ctx, failedLog)
		log.WithError(err).WithField("batch_id", batchID).Error("import failed")
		return nil, fmt.Errorf("failed to persist donations: %w", err)
	}

	result.SuccessRows = outcome.Inserted
	result.SkippedRows = len(outcome.Skipped)
	result.Warnings = append(result.Warnings, outcome.Skipped...)

	log.WithFields(logrus.Fields{
		"batch_id":     batchID,
		"success_rows": result.SuccessRows,
		"failed_rows":  result.FailedRows,
		"skipped_rows": result.SkippedRows,
	}).Info("import finished")

	return result, nil
}

// RollbackBatch soft-deletes every donation created by a batch and marks its
// log rolled back. It refuses to touch unknown or already-rolled-back batches,
// and only ever affects donations tagged with the exact batch id.
func (s *ImportService) RollbackBatch(ctx context.Context, batchID string, deletedBy *int) (int64, error) {
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"event":    "donation_import_rollback",
		"batch_id": batchID,
		"user_id":  derefInt(deletedBy),
	})
	log.Info("rollback started")

	logRecord, err := s.Repo.GetImportLogByBatchID(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if logRecord == nil {
		log.Warn("rollback rejected: batch not found")
		return 0, ErrBatchNotFound
	}
	if logRecord.Status == StatusRolledBack {
		log.Warn("rollback rejected: already rolled back")
		return 0, ErrBatchAlreadyRolledBk
	}

	deleted, err := s.Repo.SoftDeleteByBatchID(ctx, batchID, deletedBy)
	if err != nil {
		log.WithError(err).Error("rollback failed")
		return 0, err
	}

	log.WithField("deleted_rows", deleted).Info("rollback finished")
	return deleted, nil
}

// classifyRow validates a single parsed row against the pre-loaded master data
// and returns either a persistable candidate or a row error. It performs no I/O
// and makes no duplicate decision (that happens under lock during persistence).
func (s *ImportService) classifyRow(
	row ImportDonationRequest,
	donors map[string]DonaturRef,
	categories map[string]int,
) (*ImportCandidate, *DonationImportError) {
	fail := func(msg string) *DonationImportError {
		return &DonationImportError{Row: row.Row, Message: msg}
	}

	if row.DonaturID == "" {
		return nil, fail("ID Donatur is required")
	}
	if row.CategoryName == "" {
		return nil, fail("Jenis Sumbangan is required")
	}

	amount, err := parseAmount(row.Amount)
	if err != nil {
		return nil, fail(err.Error())
	}
	if amount < 0 {
		return nil, fail("amount must be greater than or equal to 0")
	}

	donor, ok := donors[row.DonaturID]
	if !ok {
		return nil, fail(fmt.Sprintf("Donatur %s not found", row.DonaturID))
	}

	categoryID, ok := categories[normalizeCategoryKey(row.CategoryName)]
	if !ok {
		return nil, fail(fmt.Sprintf("Donation category '%s' not found", row.CategoryName))
	}

	return &ImportCandidate{
		Row:      row.Row,
		DedupKey: dedupKey(donor.ID, categoryID, row.Note, amount),
		Donation: ImportedDonation{
			DonaturID:          donor.ID,
			DonaturGroupID:     donor.DonaturGroupID, // enriched from donor, may be NULL
			AreaID:             nil,                   // not present in the import sheet
			DonationCategoryID: categoryID,
			Currency:           defaultImportCurrency,
			Amount:             amount,
			OtherItems:         row.Note,
		},
	}, nil
}

func (s *ImportService) buildLog(batchID string, opts ImportOptions, meta headerMeta, result *DonationImportResult) *ImportLog {
	return &ImportLog{
		BatchID:        batchID,
		Filename:       opts.Filename,
		KomisariatID:   meta.KomisariatID,
		KomisariatName: meta.KomisariatName,
		Period:         meta.Period,
		UploadedBy:     opts.UploadedBy,
		UploadedAt:     time.Now(),
		TotalRows:      result.TotalRows,
		SuccessRows:    result.SuccessRows,
		FailedRows:     result.FailedRows,
		SkippedRows:    result.SkippedRows,
		Status:         result.Status,
	}
}

func distinctDonaturIDs(rows []ImportDonationRequest) []string {
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.DonaturID == "" {
			continue
		}
		if _, ok := seen[row.DonaturID]; ok {
			continue
		}
		seen[row.DonaturID] = struct{}{}
		out = append(out, row.DonaturID)
	}
	return out
}

func importStatus(dryRun bool, failed int) string {
	switch {
	case dryRun:
		return StatusDryRun
	case failed > 0:
		return StatusCompletedWithError
	default:
		return StatusCompleted
	}
}

// fileLockKey derives a stable advisory-lock key from the file's identity
// (komisariat + period + filename), so two simultaneous uploads of the same
// file serialise while unrelated imports run concurrently.
func fileLockKey(meta headerMeta, filename string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(meta.KomisariatID + "|" + meta.Period + "|" + filename))
	return int64(h.Sum64())
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
