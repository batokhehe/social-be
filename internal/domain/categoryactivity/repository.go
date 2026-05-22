package categoryactivity

import (
	"context"
	"fmt"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*CategoryActivity, error)
	GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]CategoryActivity, int64, error)
	GetByID(ctx context.Context, id int) (*CategoryActivity, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*CategoryActivity, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type categoryActivityModel struct {
	ID          int        `gorm:"column:id"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (categoryActivityModel) TableName() string {
	return "category_activities"
}

func toEntity(row categoryActivityModel) CategoryActivity {
	return CategoryActivity{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Status:      row.Status,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
		DeletedBy:   row.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*CategoryActivity, error) {
	row := categoryActivityModel{
		Name:        req.Name,
		Description: req.Description,
		Status:      defaultString(req.Status, "active"),
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	if err := r.DB.WithContext(ctx).Create(&row).Error; err != nil {
		r.Logger.WithError(err).Error("Create CategoryActivity failed")
		return nil, fmt.Errorf("failed create category activity: %w", err)
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]CategoryActivity, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&categoryActivityModel{}).Where("deleted_at IS NULL")
	baseQuery = query.ApplyFilters(baseQuery, filters)
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count CategoryActivity failed")
		return nil, 0, fmt.Errorf("failed count category activity: %w", err)
	}

	var rows []categoryActivityModel
	if err := baseQuery.Order("id ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated CategoryActivity failed")
		return nil, 0, fmt.Errorf("failed get paginated category activity: %w", err)
	}

	items := make([]CategoryActivity, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*CategoryActivity, error) {
	var row categoryActivityModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, err
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*CategoryActivity, error) {
	tx := r.DB.WithContext(ctx).Model(&categoryActivityModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"status":      defaultString(req.Status, "active"),
		"updated_by":  actorID,
		"updated_at":  time.Now(),
	})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("Update CategoryActivity failed")
		return nil, fmt.Errorf("failed update category activity: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&categoryActivityModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"status":     "inactive",
		"deleted_by": actorID,
		"deleted_at": time.Now(),
		"updated_by": actorID,
		"updated_at": time.Now(),
	})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("SoftDelete CategoryActivity failed")
		return fmt.Errorf("failed soft delete category activity: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func defaultString(val, def string) string {
	if val == "" {
		return def
	}

	return val
}
