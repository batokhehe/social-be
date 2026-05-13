package religion

import (
	"context"
	"fmt"
	"social-be/internal/pkg/pagination"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*Religion, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]Religion, int64, error)
	GetByID(ctx context.Context, id int) (*Religion, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Religion, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type religionModel struct {
	ID          int        `gorm:"column:id"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (religionModel) TableName() string {
	return "religions"
}

func toEntity(row religionModel) Religion {
	return Religion{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Status:      row.Status,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
		DeletedBy:   row.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*Religion, error) {
	row := religionModel{Name: req.Name, Description: req.Description, Status: defaultString(req.Status, "active"), CreatedBy: &actorID, UpdatedBy: &actorID}
	if err := r.DB.WithContext(ctx).Create(&row).Error; err != nil {
		r.Logger.WithError(err).Error("Create Religion failed")
		return nil, fmt.Errorf("failed create religion: %w", err)
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]Religion, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&religionModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed count religion: %w", err)
	}

	var rows []religionModel
	if err := baseQuery.Order("id ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed get paginated religion: %w", err)
	}

	items := make([]Religion, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Religion, error) {
	var row religionModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, err
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Religion, error) {
	tx := r.DB.WithContext(ctx).Model(&religionModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"name": req.Name, "description": req.Description, "status": defaultString(req.Status, "active"), "updated_by": actorID, "updated_at": time.Now(),
	})
	if tx.Error != nil {
		return nil, fmt.Errorf("failed update religion: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&religionModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"status": "inactive", "deleted_by": actorID, "deleted_at": time.Now(), "updated_by": actorID, "updated_at": time.Now(),
	})
	if tx.Error != nil {
		return fmt.Errorf("failed soft delete religion: %w", tx.Error)
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
