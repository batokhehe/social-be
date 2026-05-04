package levelarea

import (
	"time"

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

func (r *Repository) Create(req CreateRequest, actorID int) (*LevelArea, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	item := levelAreaModel{
		Level:       req.Level,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	if err := r.DB.Create(&item).Error; err != nil {
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *Repository) GetAll() ([]LevelArea, error) {
	var rows []levelAreaModel
	if err := r.DB.Where("deleted_at IS NULL").Order("level ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]LevelArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, nil
}

func (r *Repository) GetByID(id int) (*LevelArea, error) {
	var item levelAreaModel
	if err := r.DB.Where("id = ? AND deleted_at IS NULL", id).First(&item).Error; err != nil {
		return nil, err
	}
	out := toEntity(item)
	return &out, nil
}

func (r *Repository) Update(id int, req UpdateRequest, actorID int) (*LevelArea, error) {
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

	if err := r.DB.Model(&levelAreaModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) SoftDelete(id int, actorID int) error {
	updates := map[string]any{
		"deleted_at": gorm.Expr("NOW()"),
		"deleted_by": actorID,
		"updated_by": actorID,
		"updated_at": gorm.Expr("NOW()"),
		"status":     "inactive",
	}
	return r.DB.Model(&levelAreaModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}
