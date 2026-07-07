package masterexpensecategory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("expense category not found")
	ErrCodeExists       = errors.New("code already exists")
	ErrNameExists       = errors.New("name already exists")
)

// Sort is a validated (whitelisted) ORDER BY.
type Sort struct {
	Column string
	Order  string
}

func (s Sort) orderClause() string {
	return s.Column + " " + s.Order
}

type Repository interface {
	Create(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error)
	GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort Sort) ([]MasterExpenseCategory, int64, error)
	GetSelect(ctx context.Context) ([]SelectItem, error)
	GetByID(ctx context.Context, id int) (*MasterExpenseCategory, error)
	Update(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error)
	SoftDelete(ctx context.Context, id int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type masterExpenseCategoryModel struct {
	ID        int        `gorm:"column:id"`
	Code      string     `gorm:"column:code"`
	Name      string     `gorm:"column:name"`
	Active    bool       `gorm:"column:active"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (masterExpenseCategoryModel) TableName() string { return "master_expense_categories" }

func toEntity(row masterExpenseCategoryModel) MasterExpenseCategory {
	return MasterExpenseCategory{
		ID: row.ID, Code: row.Code, Name: row.Name, Active: row.Active,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// codeExists / nameExists enforce uniqueness against ALL records, including
// soft-deleted ones: a code/name that has ever existed can never be reused.
// The check is at the application layer (not relying on the DB constraint) so
// the caller can return a clean 409 instead of a raw constraint error.
// excludeID skips the row being updated.
func (r *GormRepository) codeExists(ctx context.Context, code string, excludeID int) (bool, error) {
	q := r.DB.WithContext(ctx).Model(&masterExpenseCategoryModel{}).
		Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepository) nameExists(ctx context.Context, name string, excludeID int) (bool, error) {
	q := r.DB.WithContext(ctx).Model(&masterExpenseCategoryModel{}).
		Where("name = ?", name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// isUniqueViolation detects a Postgres unique-constraint error (SQLSTATE 23505)
// without importing a driver package, so a race that slips past the app-level
// check is still translated to a business error instead of leaking to the API.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "sqlstate 23505") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "unique constraint")
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest) (*MasterExpenseCategory, error) {
	if exists, err := r.codeExists(ctx, req.Code, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrCodeExists
	}
	if exists, err := r.nameExists(ctx, req.Name, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrNameExists
	}

	row := masterExpenseCategoryModel{Code: req.Code, Name: req.Name, Active: *req.Active}
	if err := r.DB.WithContext(ctx).Select("code", "name", "active").Create(&row).Error; err != nil {
		// Backstop for a race that slips past the app-level checks: the only DB
		// unique constraint is on code, so translate to a clean 409.
		if isUniqueViolation(err) {
			return nil, ErrCodeExists
		}
		r.Logger.WithError(err).Error("Create MasterExpenseCategory failed")
		return nil, fmt.Errorf("failed create expense category: %w", err)
	}
	return r.GetByID(ctx, row.ID)
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters, sort Sort) ([]MasterExpenseCategory, int64, error) {
	baseQuery := r.DB.WithContext(ctx).Model(&masterExpenseCategoryModel{}).Where("deleted_at IS NULL")
	baseQuery = query.ApplyFilters(baseQuery, filters)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count MasterExpenseCategory failed")
		return nil, 0, fmt.Errorf("failed count expense category: %w", err)
	}

	var rows []masterExpenseCategoryModel
	if err := baseQuery.
		Order(sort.orderClause()).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated MasterExpenseCategory failed")
		return nil, 0, fmt.Errorf("failed get paginated expense category: %w", err)
	}

	items := make([]MasterExpenseCategory, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, total, nil
}

func (r *GormRepository) GetSelect(ctx context.Context) ([]SelectItem, error) {
	var items []SelectItem
	if err := r.DB.WithContext(ctx).
		Model(&masterExpenseCategoryModel{}).
		Select("id", "code", "name").
		Where("active = true AND deleted_at IS NULL").
		Order("name ASC").
		Find(&items).Error; err != nil {
		r.Logger.WithError(err).Error("GetSelect MasterExpenseCategory failed")
		return nil, fmt.Errorf("failed get select expense category: %w", err)
	}
	return items, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterExpenseCategory, error) {
	var row masterExpenseCategoryModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	item := toEntity(row)
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest) (*MasterExpenseCategory, error) {
	if exists, err := r.codeExists(ctx, req.Code, id); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrCodeExists
	}
	if exists, err := r.nameExists(ctx, req.Name, id); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrNameExists
	}

	result := r.DB.WithContext(ctx).Model(&masterExpenseCategoryModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"code":       req.Code,
			"name":       req.Name,
			"active":     *req.Active,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, ErrCodeExists
		}
		return nil, fmt.Errorf("failed update expense category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrCategoryNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int) error {
	result := r.DB.WithContext(ctx).Model(&masterExpenseCategoryModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}
