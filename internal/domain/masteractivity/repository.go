package masteractivity

import (
	"context"
	"fmt"
	"social-be/internal/pkg/pagination"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*MasterActivity, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]MasterActivity, int64, error)
	GetByID(ctx context.Context, id int) (*MasterActivity, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterActivity, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type masterActivityModel struct {
	ID                 int        `gorm:"column:id"`
	CategoryActivityID int        `gorm:"column:category_activity_id"`
	Name               string     `gorm:"column:name"`
	Target             int        `gorm:"column:target"`
	Description        string     `gorm:"column:description"`
	Status             string     `gorm:"column:status"`
	CreatedBy          *int       `gorm:"column:created_by"`
	UpdatedBy          *int       `gorm:"column:updated_by"`
	DeletedBy          *int       `gorm:"column:deleted_by"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
}

func (masterActivityModel) TableName() string {
	return "master_activities"
}

func toEntity(row masterActivityModel) MasterActivity {
	return MasterActivity{ID: row.ID, CategoryActivityID: row.CategoryActivityID, Name: row.Name, Target: row.Target, Description: row.Description, Status: row.Status, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, DeletedBy: row.DeletedBy}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*MasterActivity, error) {
	row := masterActivityModel{CategoryActivityID: req.CategoryActivityID, Name: req.Name, Target: req.Target, Description: req.Description, Status: defaultString(req.Status, "active"), CreatedBy: &actorID, UpdatedBy: &actorID}
	if err := r.DB.WithContext(ctx).Create(&row).Error; err != nil {
		r.Logger.WithError(err).Error("Create MasterActivity failed")
		return nil, fmt.Errorf("failed create master activity: %w", err)
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]MasterActivity, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&masterActivityModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed count master activity: %w", err)
	}
	var rows []masterActivityModel
	if err := baseQuery.Order("category_activity_id ASC, id ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed get paginated master activity: %w", err)
	}
	items := make([]MasterActivity, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterActivity, error) {
	var row masterActivityModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, err
	}
	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterActivity, error) {
	tx := r.DB.WithContext(ctx).Model(&masterActivityModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"category_activity_id": req.CategoryActivityID, "name": req.Name, "target": req.Target, "description": req.Description, "status": defaultString(req.Status, "active"), "updated_by": actorID, "updated_at": time.Now(),
	})
	if tx.Error != nil {
		return nil, fmt.Errorf("failed update master activity: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&masterActivityModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"status": "inactive", "deleted_by": actorID, "deleted_at": time.Now(), "updated_by": actorID, "updated_at": time.Now(),
	})
	if tx.Error != nil {
		return fmt.Errorf("failed soft delete master activity: %w", tx.Error)
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
