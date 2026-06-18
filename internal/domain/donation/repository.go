package donation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, donation *Donation) (*Donation, error)
	GetAll(ctx context.Context) ([]Donation, error)
	GetByID(ctx context.Context, id int) (*Donation, error)
	Update(ctx context.Context, donation *Donation) (*Donation, error)
	Delete(ctx context.Context, id int) error

	// Import-related bulk operations. These are designed to keep the number of
	// queries constant regardless of row count (no N+1).

	// FindDonaturRefsByIDs resolves many master_donaturs rows in one query,
	// keyed by their business id ("ID Donatur"). Missing/soft-deleted donors are
	// simply absent from the map.
	FindDonaturRefsByIDs(ctx context.Context, donaturIDs []string) (map[string]DonaturRef, error)
	// LoadCategoryIndex loads the (small) active category master into a map of
	// lower-cased, trimmed name -> id for case-insensitive lookups.
	LoadCategoryIndex(ctx context.Context) (map[string]int, error)
	// LoadExistingDedupKeys returns the dedup keys of live donations for the
	// given donors. Used by the dry-run path (no persistence, race irrelevant).
	LoadExistingDedupKeys(ctx context.Context, donaturIDs []int) (map[string]struct{}, error)
	// PersistImportLocked persists an import race-free. It takes a transaction
	// advisory lock keyed on the file identity (serialising concurrent imports
	// of the same file), re-checks duplicates inside that lock, inserts the
	// surviving rows in batches, persists the failed rows, and writes the audit
	// log -- all in one transaction so data and log stay consistent.
	PersistImportLocked(ctx context.Context, lockKey int64, donorPKs []int, candidates []ImportCandidate, errorRows []ImportError, log *ImportLog) (PersistOutcome, error)
	// CreateImportLog records an import that produced no inserted rows (e.g. a
	// fully-failed file) for audit purposes.
	CreateImportLog(ctx context.Context, log *ImportLog) error
	// GetImportLogByBatchID returns the audit log for a batch, or (nil, nil) if
	// no such batch exists.
	GetImportLogByBatchID(ctx context.Context, batchID string) (*ImportLog, error)
	// SoftDeleteByBatchID soft-deletes every live donation belonging to a batch
	// (rollback), stamping deleted_by, and marks the corresponding log as rolled
	// back (with rolled_back_by / rolled_back_at). Returns the number of
	// donations soft-deleted.
	SoftDeleteByBatchID(ctx context.Context, batchID string, deletedBy *int) (int64, error)

	// --- History & audit reads ---

	// GetImportLogs returns a filtered, sorted, paginated page of import logs.
	GetImportLogs(ctx context.Context, filters query.Filters, page pagination.Query, sort ImportHistorySort) ([]ImportLog, int64, error)
	// GetImportErrors returns the persisted failed rows for a batch.
	GetImportErrors(ctx context.Context, batchID string) ([]ImportError, error)
	// GetDonationsByBatchID returns the (live) donations created by a batch.
	GetDonationsByBatchID(ctx context.Context, batchID string) ([]Donation, error)
	// GetUserNames resolves user ids to display names for audit enrichment.
	GetUserNames(ctx context.Context, ids []int) (map[int]string, error)
}

// ImportCandidate is a validated row ready for persistence, paired with its
// duplicate key and source row number for accurate reporting.
type ImportCandidate struct {
	Row      int
	DedupKey string
	Donation ImportedDonation
}

// PersistOutcome reports what actually happened inside the locked transaction.
type PersistOutcome struct {
	Inserted int
	Skipped  []DonationImportError
}

// DonaturRef is the minimal master_donaturs projection the import needs: the
// business id, the primary key used as donations.donatur_id, and the donor's
// group (enriched into the donation when present).
type DonaturRef struct {
	DonaturID      string `gorm:"column:id_donatur"`
	ID             int    `gorm:"column:id"`
	DonaturGroupID *int   `gorm:"column:id_group_donatur"`
}

// ImportedDonation is the insert payload for an Excel-imported donation. Group
// and area are pointers so they can be stored as NULL.
type ImportedDonation struct {
	DonaturID          int     `gorm:"column:donatur_id"`
	DonaturGroupID     *int    `gorm:"column:donatur_group_id"`
	AreaID             *int    `gorm:"column:area_id"`
	DonationCategoryID int     `gorm:"column:donation_category_id"`
	Currency           string  `gorm:"column:currency"`
	Amount             float64 `gorm:"column:amount"`
	OtherItems         string  `gorm:"column:other_items"`
	ImportBatchID      string  `gorm:"column:import_batch_id"`
}

func (ImportedDonation) TableName() string { return "donations" }

type GormRepository struct {
	DB *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{DB: db}
}

func (r *GormRepository) Create(ctx context.Context, donation *Donation) (*Donation, error) {
	if err := r.DB.WithContext(ctx).Create(donation).Error; err != nil {
		return nil, err
	}
	return donation, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]Donation, error) {
	var donations []Donation
	if err := r.DB.WithContext(ctx).Order("id ASC").Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Donation, error) {
	var donation Donation
	if err := r.DB.WithContext(ctx).First(&donation, id).Error; err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *GormRepository) Update(ctx context.Context, donation *Donation) (*Donation, error) {
	if err := r.DB.WithContext(ctx).Save(donation).Error; err != nil {
		return nil, err
	}
	return donation, nil
}

func (r *GormRepository) Delete(ctx context.Context, id int) error {
	if err := r.DB.WithContext(ctx).Delete(&Donation{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormRepository) FindDonaturRefsByIDs(ctx context.Context, donaturIDs []string) (map[string]DonaturRef, error) {
	out := make(map[string]DonaturRef, len(donaturIDs))
	if len(donaturIDs) == 0 {
		return out, nil
	}

	var refs []DonaturRef
	err := r.DB.WithContext(ctx).
		Table("master_donaturs").
		Select("id_donatur", "id", "id_group_donatur").
		Where("id_donatur IN ? AND deleted_at IS NULL", donaturIDs).
		Find(&refs).Error
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		out[ref.DonaturID] = ref
	}
	return out, nil
}

func (r *GormRepository) LoadCategoryIndex(ctx context.Context) (map[string]int, error) {
	var rows []struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	err := r.DB.WithContext(ctx).
		Table("master_donation_categories").
		Select("id", "name").
		Where("deleted_at IS NULL").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[normalizeCategoryKey(row.Name)] = row.ID
	}
	return out, nil
}

func (r *GormRepository) LoadExistingDedupKeys(ctx context.Context, donaturIDs []int) (map[string]struct{}, error) {
	return loadDedupKeys(r.DB.WithContext(ctx), donaturIDs)
}

// loadDedupKeys builds the set of dedup keys for live donations of the given
// donors using the provided DB/transaction handle.
func loadDedupKeys(db *gorm.DB, donaturIDs []int) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if len(donaturIDs) == 0 {
		return out, nil
	}

	var rows []struct {
		DonaturID          int     `gorm:"column:donatur_id"`
		DonationCategoryID int     `gorm:"column:donation_category_id"`
		OtherItems         string  `gorm:"column:other_items"`
		Amount             float64 `gorm:"column:amount"`
	}
	err := db.
		Table("donations").
		Select("donatur_id", "donation_category_id", "other_items", "amount").
		Where("donatur_id IN ? AND deleted_at IS NULL", donaturIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		out[dedupKey(row.DonaturID, row.DonationCategoryID, row.OtherItems, row.Amount)] = struct{}{}
	}
	return out, nil
}

func (r *GormRepository) PersistImportLocked(ctx context.Context, lockKey int64, donorPKs []int, candidates []ImportCandidate, errorRows []ImportError, log *ImportLog) (PersistOutcome, error) {
	var outcome PersistOutcome
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialise concurrent imports of the same file. The lock is released
		// automatically at the end of the transaction.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey).Error; err != nil {
			return err
		}

		// Authoritative duplicate snapshot, taken under the lock so a concurrent
		// import committed between parse and now is also seen.
		existing, err := loadDedupKeys(tx, donorPKs)
		if err != nil {
			return err
		}

		toInsert := make([]ImportedDonation, 0, len(candidates))
		skipped := make([]DonationImportError, 0)
		seen := make(map[string]struct{}, len(candidates))
		for _, c := range candidates {
			if _, dup := existing[c.DedupKey]; dup {
				skipped = append(skipped, DonationImportError{Row: c.Row, Message: "Donation already exists"})
				continue
			}
			if _, dup := seen[c.DedupKey]; dup {
				skipped = append(skipped, DonationImportError{Row: c.Row, Message: "Donation already exists"})
				continue
			}
			seen[c.DedupKey] = struct{}{}
			toInsert = append(toInsert, c.Donation)
		}

		if len(toInsert) > 0 {
			if err := tx.CreateInBatches(&toInsert, importInsertBatchSize).Error; err != nil {
				return err
			}
		}

		if len(errorRows) > 0 {
			if err := tx.CreateInBatches(&errorRows, importInsertBatchSize).Error; err != nil {
				return err
			}
		}

		log.SuccessRows = len(toInsert)
		log.SkippedRows = len(skipped)
		if err := tx.Create(log).Error; err != nil {
			return err
		}

		outcome = PersistOutcome{Inserted: len(toInsert), Skipped: skipped}
		return nil
	})
	if err != nil {
		return PersistOutcome{}, err
	}
	return outcome, nil
}

func (r *GormRepository) CreateImportLog(ctx context.Context, log *ImportLog) error {
	return r.DB.WithContext(ctx).Create(log).Error
}

func (r *GormRepository) GetImportLogByBatchID(ctx context.Context, batchID string) (*ImportLog, error) {
	var log ImportLog
	err := r.DB.WithContext(ctx).Where("batch_id = ?", batchID).Take(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *GormRepository) SoftDeleteByBatchID(ctx context.Context, batchID string, deletedBy *int) (int64, error) {
	var deleted int64
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Direct column update (rather than GORM's Delete) so we can stamp
		// deleted_by alongside deleted_at in a single statement. The model's
		// soft-delete config keeps these rows out of subsequent reads.
		res := tx.Model(&Donation{}).
			Where("import_batch_id = ? AND deleted_at IS NULL", batchID).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			})
		if res.Error != nil {
			return res.Error
		}
		deleted = res.RowsAffected

		return tx.Model(&ImportLog{}).
			Where("batch_id = ?", batchID).
			Updates(map[string]any{
				"status":         StatusRolledBack,
				"rolled_back_by": deletedBy,
				"rolled_back_at": time.Now(),
			}).Error
	})
	return deleted, err
}

func (r *GormRepository) GetImportLogs(ctx context.Context, filters query.Filters, page pagination.Query, sort ImportHistorySort) ([]ImportLog, int64, error) {
	base := r.DB.WithContext(ctx).Model(&ImportLog{})
	base = query.ApplyFilters(base, filters)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count import logs: %w", err)
	}

	var logs []ImportLog
	if err := base.
		Order(sort.OrderClause()).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch import logs: %w", err)
	}
	return logs, total, nil
}

func (r *GormRepository) GetImportErrors(ctx context.Context, batchID string) ([]ImportError, error) {
	var rows []ImportError
	if err := r.DB.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Order("row_number ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch import errors: %w", err)
	}
	return rows, nil
}

func (r *GormRepository) GetDonationsByBatchID(ctx context.Context, batchID string) ([]Donation, error) {
	var donations []Donation
	// Soft-deleted rows are auto-excluded by the model's gorm.DeletedAt.
	if err := r.DB.WithContext(ctx).
		Where("import_batch_id = ?", batchID).
		Order("id ASC").
		Find(&donations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch batch donations: %w", err)
	}
	return donations, nil
}

func (r *GormRepository) GetUserNames(ctx context.Context, ids []int) (map[int]string, error) {
	out := make(map[int]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	var rows []struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	if err := r.DB.WithContext(ctx).
		Table("users").
		Select("id", "name").
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user names: %w", err)
	}
	for _, row := range rows {
		out[row.ID] = row.Name
	}
	return out, nil
}

// importInsertBatchSize bounds how many donation rows are sent per INSERT.
const importInsertBatchSize = 500

// normalizeCategoryKey makes category names comparable: trimmed + lower-cased.
func normalizeCategoryKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// dedupKey builds the duplicate-detection key (same donor, category, note,
// amount). Amount is fixed to 2 decimals to match NUMERIC(15,2).
func dedupKey(donaturID, categoryID int, note string, amount float64) string {
	return fmt.Sprintf("%d|%d|%.2f|%s", donaturID, categoryID, amount, note)
}
