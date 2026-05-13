package attributevolunteer

import (
	"context"
	"fmt"
	"social-be/internal/pkg/pagination"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*AttributeVolunteer, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]AttributeVolunteer, int64, error)
	GetByID(ctx context.Context, id int) (*AttributeVolunteer, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*AttributeVolunteer, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type attributeVolunteerModel struct {
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

func (attributeVolunteerModel) TableName() string {
	return "attribute_volunteers"
}

func toEntity(row attributeVolunteerModel) AttributeVolunteer {
	return AttributeVolunteer{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Status:      row.Status,
		CreatedBy:   row.CreatedBy,
		UpdatedBy:   row.UpdatedBy,
		DeletedBy:   row.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*AttributeVolunteer, error) {
	row := attributeVolunteerModel{
		Name:        req.Name,
		Description: req.Description,
		Status:      defaultString(req.Status, "active"),
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	if err := r.DB.WithContext(ctx).Create(&row).Error; err != nil {
		r.Logger.WithError(err).Error("Create AttributeVolunteer failed")
		return nil, fmt.Errorf("failed create attribute volunteer: %w", err)
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]AttributeVolunteer, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&attributeVolunteerModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count AttributeVolunteer failed")
		return nil, 0, fmt.Errorf("failed count attribute volunteer: %w", err)
	}

	var rows []attributeVolunteerModel
	if err := baseQuery.Order("id ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated AttributeVolunteer failed")
		return nil, 0, fmt.Errorf("failed get paginated attribute volunteer: %w", err)
	}

	items := make([]AttributeVolunteer, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*AttributeVolunteer, error) {
	var row attributeVolunteerModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, err
	}

	out := toEntity(row)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*AttributeVolunteer, error) {
	tx := r.DB.WithContext(ctx).Model(&attributeVolunteerModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"name":        req.Name,
		"description": req.Description,
		"status":      defaultString(req.Status, "active"),
		"updated_by":  actorID,
		"updated_at":  time.Now(),
	})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("Update AttributeVolunteer failed")
		return nil, fmt.Errorf("failed update attribute volunteer: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&attributeVolunteerModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"status":     "inactive",
		"deleted_by": actorID,
		"deleted_at": time.Now(),
		"updated_by": actorID,
		"updated_at": time.Now(),
	})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("SoftDelete AttributeVolunteer failed")
		return fmt.Errorf("failed soft delete attribute volunteer: %w", tx.Error)
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
