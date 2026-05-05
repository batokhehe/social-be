package levelarea

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*LevelArea, error)
	GetAll(ctx context.Context) ([]LevelArea, error)
	GetByID(ctx context.Context, id int) (*LevelArea, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelArea, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{
		DB:     db,
		Logger: logger,
	}
}

type levelAreaModel struct {
	ID          int        `gorm:"column:id"`
	Level       int        `gorm:"column:level"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (levelAreaModel) TableName() string {
	return "level_areas"
}

func toEntity(item levelAreaModel) LevelArea {
	return LevelArea{
		ID:          item.ID,
		Level:       item.Level,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*LevelArea, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	level := req.Level
	if level == 0 {
		level = 1
	}

	item := levelAreaModel{
		Level:       level,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.BelowLevelAreaID != nil {
			var anchor levelAreaModel
			if err := tx.
				Where("id = ? AND deleted_at IS NULL", *req.BelowLevelAreaID).
				First(&anchor).Error; err != nil {
				return fmt.Errorf("anchor not found: %w", err)
			}

			insertLevel := anchor.Level + 1

			if err := tx.Model(&levelAreaModel{}).
				Where("deleted_at IS NULL AND level >= ?", insertLevel).
				Update("level", gorm.Expr("level + 1")).Error; err != nil {
				return fmt.Errorf("failed shift level: %w", err)
			}

			item.Level = insertLevel
		}

		result := tx.Create(&item)
		if result.Error != nil {
			return fmt.Errorf("failed insert level area: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("insert failed: no rows affected")
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create LevelArea failed")
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]LevelArea, error) {
	var rows []levelAreaModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("level ASC").
		Find(&rows).Error

	if err != nil {
		r.Logger.WithError(err).Error("GetAll LevelArea failed")
		return nil, fmt.Errorf("failed get all level area: %w", err)
	}

	items := make([]LevelArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*LevelArea, error) {
	var item levelAreaModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get level area by id: %w", err)
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*LevelArea, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	updates := map[string]any{
		"level":       req.Level,
		"name":        req.Name,
		"description": req.Description,
		"status":      status,
		"updated_by":  actorID,
		"updated_at":  time.Now(),
	}

	tx := r.DB.WithContext(ctx).
		Model(&levelAreaModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)

	if tx.Error != nil {
		return nil, fmt.Errorf("failed update level area: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	updates := map[string]any{
		"deleted_at": time.Now(),
		"deleted_by": actorID,
		"updated_by": actorID,
		"updated_at": time.Now(),
		"status":     "inactive",
	}

	tx := r.DB.WithContext(ctx).
		Model(&levelAreaModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)

	if tx.Error != nil {
		return fmt.Errorf("failed soft delete: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
