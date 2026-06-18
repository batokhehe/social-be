package donation

import (
	"context"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
)

// ImportHistoryService serves the read-only import history & audit endpoints.
type ImportHistoryService struct {
	Repo Repository
}

func NewImportHistoryService(repo Repository) *ImportHistoryService {
	return &ImportHistoryService{Repo: repo}
}

// List returns a filtered, sorted, paginated page of import history, enriched
// with uploader / rollback-actor names.
func (s *ImportHistoryService) List(ctx context.Context, filters query.Filters, page pagination.Query, sort ImportHistorySort) ([]ImportHistoryItem, pagination.Meta, error) {
	logs, total, err := s.Repo.GetImportLogs(ctx, filters, page, sort)
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	names, err := s.Repo.GetUserNames(ctx, collectUserIDs(logs))
	if err != nil {
		return nil, pagination.Meta{}, err
	}

	items := make([]ImportHistoryItem, 0, len(logs))
	for i := range logs {
		items = append(items, toHistoryItem(logs[i], names))
	}
	return items, pagination.NewMeta(page.Page, page.Limit, total), nil
}

// Detail returns the single-batch detail view. ErrBatchNotFound if unknown.
func (s *ImportHistoryService) Detail(ctx context.Context, batchID string) (*ImportDetailResponse, error) {
	log, err := s.requireLog(ctx, batchID)
	if err != nil {
		return nil, err
	}

	names, err := s.Repo.GetUserNames(ctx, collectUserIDs([]ImportLog{*log}))
	if err != nil {
		return nil, err
	}

	resp := &ImportDetailResponse{
		BatchID:        log.BatchID,
		Filename:       log.Filename,
		UploadedBy:     log.UploadedBy,
		UploadedByName: nameOf(names, log.UploadedBy),
		UploadedAt:     log.UploadedAt,
		KomisariatID:   log.KomisariatID,
		KomisariatName: log.KomisariatName,
		Period:         log.Period,
		Status:         log.Status,
		Summary: ImportSummary{
			TotalRows:   log.TotalRows,
			SuccessRows: log.SuccessRows,
			FailedRows:  log.FailedRows,
			SkippedRows: log.SkippedRows,
		},
		Rollback: toRollbackInfo(*log, names),
	}
	return resp, nil
}

// Errors returns the persisted failed rows for a batch. ErrBatchNotFound if unknown.
func (s *ImportHistoryService) Errors(ctx context.Context, batchID string) ([]ImportError, error) {
	if _, err := s.requireLog(ctx, batchID); err != nil {
		return nil, err
	}
	return s.Repo.GetImportErrors(ctx, batchID)
}

// Donations returns the donations created by a batch. ErrBatchNotFound if unknown.
func (s *ImportHistoryService) Donations(ctx context.Context, batchID string) ([]Donation, error) {
	if _, err := s.requireLog(ctx, batchID); err != nil {
		return nil, err
	}
	return s.Repo.GetDonationsByBatchID(ctx, batchID)
}

// Rollback returns who/when a batch was rolled back. ErrBatchNotFound if unknown.
func (s *ImportHistoryService) Rollback(ctx context.Context, batchID string) (*RollbackInfo, error) {
	log, err := s.requireLog(ctx, batchID)
	if err != nil {
		return nil, err
	}
	names, err := s.Repo.GetUserNames(ctx, collectUserIDs([]ImportLog{*log}))
	if err != nil {
		return nil, err
	}
	info := toRollbackInfo(*log, names)
	return &info, nil
}

// ExportErrors builds an .xlsx error report for a batch. Returns the file bytes
// and a suggested download filename. ErrBatchNotFound if unknown.
func (s *ImportHistoryService) ExportErrors(ctx context.Context, batchID string) ([]byte, string, error) {
	log, err := s.requireLog(ctx, batchID)
	if err != nil {
		return nil, "", err
	}
	rows, err := s.Repo.GetImportErrors(ctx, batchID)
	if err != nil {
		return nil, "", err
	}
	return buildErrorReport(*log, rows)
}

func (s *ImportHistoryService) requireLog(ctx context.Context, batchID string) (*ImportLog, error) {
	log, err := s.Repo.GetImportLogByBatchID(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrBatchNotFound
	}
	return log, nil
}

func toHistoryItem(log ImportLog, names map[int]string) ImportHistoryItem {
	return ImportHistoryItem{
		BatchID:          log.BatchID,
		Filename:         log.Filename,
		UploadedBy:       log.UploadedBy,
		UploadedByName:   nameOf(names, log.UploadedBy),
		UploadedAt:       log.UploadedAt,
		KomisariatID:     log.KomisariatID,
		KomisariatName:   log.KomisariatName,
		Period:           log.Period,
		TotalRows:        log.TotalRows,
		SuccessRows:      log.SuccessRows,
		FailedRows:       log.FailedRows,
		SkippedRows:      log.SkippedRows,
		Status:           log.Status,
		RolledBackBy:     log.RolledBackBy,
		RolledBackByName: nameOf(names, log.RolledBackBy),
		RolledBackAt:     log.RolledBackAt,
	}
}

func toRollbackInfo(log ImportLog, names map[int]string) RollbackInfo {
	return RollbackInfo{
		RolledBack:       log.Status == StatusRolledBack,
		Status:           log.Status,
		RolledBackBy:     log.RolledBackBy,
		RolledBackByName: nameOf(names, log.RolledBackBy),
		RolledBackAt:     log.RolledBackAt,
	}
}

func collectUserIDs(logs []ImportLog) []int {
	seen := make(map[int]struct{})
	var ids []int
	add := func(id *int) {
		if id == nil {
			return
		}
		if _, ok := seen[*id]; ok {
			return
		}
		seen[*id] = struct{}{}
		ids = append(ids, *id)
	}
	for i := range logs {
		add(logs[i].UploadedBy)
		add(logs[i].RolledBackBy)
	}
	return ids
}

func nameOf(names map[int]string, id *int) string {
	if id == nil {
		return ""
	}
	return names[*id]
}
