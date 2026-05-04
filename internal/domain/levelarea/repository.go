package levelarea

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(req CreateRequest, actorID int) (*LevelArea, error)
	GetAll() ([]LevelArea, error)
	GetByID(id int) (*LevelArea, error)
	Update(id int, req UpdateRequest, actorID int) (*LevelArea, error)
	SoftDelete(id int, actorID int) error
}

type GormRepository struct {

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
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

func (r *GormRepository) Create(req CreateRequest, actorID int) (*LevelArea, error) {
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

	if err := r.DB.Transaction(func(tx *gorm.DB) error {
		if req.BelowLevelAreaID != nil {
			var anchor levelAreaModel
			if err := tx.Where("id = ? AND deleted_at IS NULL", *req.BelowLevelAreaID).First(&anchor).Error; err != nil {
				return err
			}

			insertLevel := anchor.Level + 1
			if err := tx.Model(&levelAreaModel{}).
				Where("deleted_at IS NULL AND level >= ?", insertLevel).
				Update("level", gorm.Expr("level + 1")).Error; err != nil {
				return err
			}
			item.Level = insertLevel
		}

		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll() ([]LevelArea, error) {
	var rows []levelAreaModel
	if err := r.DB.Select("id", "level", "name", "description", "status", "created_by", "updated_by", "deleted_by").
		Where("deleted_at IS NULL").
		Order("level ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]LevelArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, nil
}

func (r *GormRepository) GetByID(id int) (*LevelArea, error) {
	var item levelAreaModel
	if err := r.DB.Select("id", "level", "name", "description", "status", "created_by", "updated_by", "deleted_by").
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error; err != nil {
		return nil, err
	}
	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) Update(id int, req UpdateRequest, actorID int) (*LevelArea, error) {
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
		"updated_at":  gorm.Expr("NOW()"),
	}

	tx := r.DB.Model(&levelAreaModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(id)
}

func (r *GormRepository) SoftDelete(id int, actorID int) error {
	updates := map[string]any{
		"deleted_at": gorm.Expr("NOW()"),
		"deleted_by": actorID,
		"updated_by": actorID,
		"updated_at": gorm.Expr("NOW()"),
		"status":     "inactive",
	}
	tx := r.DB.Model(&levelAreaModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
