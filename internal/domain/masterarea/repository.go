package masterarea

import (
	"context"
	"errors"
	"fmt"
	"social-be/internal/pkg/pagination"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	errInvalidLevelAreaID = errors.New("level_area_id must be a valid level area id")
	errParentRequired     = errors.New("parent_id is required for selected level area")
	errParentNotAllowed   = errors.New("parent_id is not allowed for root level area")
	errInvalidParentLevel = errors.New("parent_id must reference master area from previous level area")
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*MasterArea, error)
	GetAll(ctx context.Context) ([]MasterArea, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]MasterArea, int64, error)
	GetByID(ctx context.Context, id int) (*MasterArea, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterArea, error)
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

type masterAreaModel struct {
	ID          int        `gorm:"column:id"`
	Name        string     `gorm:"column:name"`
	LevelAreaID string     `gorm:"column:level_area_id"`
	Description string     `gorm:"column:description"`
	ParentID    *int       `gorm:"column:parent_id"`
	Location    string     `gorm:"column:location"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (masterAreaModel) TableName() string {
	return "master_areas"
}

type levelAreaModel struct {
	ID    int `gorm:"column:id"`
	Level int `gorm:"column:level"`
}

func (levelAreaModel) TableName() string {
	return "level_areas"
}

func toEntity(item masterAreaModel) MasterArea {
	return MasterArea{
		ID:          item.ID,
		Name:        item.Name,
		LevelAreaID: item.LevelAreaID,
		Description: item.Description,
		ParentID:    item.ParentID,
		Location:    item.Location,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*MasterArea, error) {
	item := masterAreaModel{
		Name:        req.Name,
		LevelAreaID: req.LevelAreaID,
		Description: req.Description,
		ParentID:    req.ParentID,
		Location:    req.Location,
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateParent(ctx, tx, req.LevelAreaID, req.ParentID, nil); err != nil {
			return err
		}

		result := tx.Create(&item)
		if result.Error != nil {
			return fmt.Errorf("failed insert master area: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("insert failed: no rows affected")
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create MasterArea failed")
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]MasterArea, error) {
	var rows []masterAreaModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error

	if err != nil {
		r.Logger.WithError(err).Error("GetAll MasterArea failed")
		return nil, fmt.Errorf("failed get all master area: %w", err)
	}

	items := make([]MasterArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]MasterArea, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&masterAreaModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count MasterArea failed")
		return nil, 0, fmt.Errorf("failed count master area: %w", err)
	}

	var rows []masterAreaModel
	if err := baseQuery.
		Order("id ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated MasterArea failed")
		return nil, 0, fmt.Errorf("failed get paginated master area: %w", err)
	}

	items := make([]MasterArea, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterArea, error) {
	var item masterAreaModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get master area by id: %w", err)
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterArea, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing masterAreaModel
		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			Take(&existing).Error; err != nil {
			return err
		}

		if err := validateParent(ctx, tx, req.LevelAreaID, req.ParentID, &id); err != nil {
			return err
		}

		updates := map[string]any{
			"name":          req.Name,
			"level_area_id": req.LevelAreaID,
			"description":   req.Description,
			"parent_id":     req.ParentID,
			"location":      req.Location,
			"updated_by":    actorID,
			"updated_at":    time.Now(),
		}

		result := tx.Model(&masterAreaModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed update master area: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update MasterArea failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	updates := map[string]any{
		"deleted_at": time.Now(),
		"deleted_by": actorID,
		"updated_by": actorID,
		"updated_at": time.Now(),
	}

	tx := r.DB.WithContext(ctx).
		Model(&masterAreaModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)

	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("SoftDelete MasterArea failed")
		return fmt.Errorf("failed soft delete master area: %w", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func validateParent(ctx context.Context, tx *gorm.DB, levelAreaID string, parentID *int, currentID *int) error {
	levelAreaIntID, err := strconv.Atoi(levelAreaID)
	if err != nil {
		return errInvalidLevelAreaID
	}

	var selected levelAreaModel
	if err := tx.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", levelAreaIntID).
		Take(&selected).Error; err != nil {
		return fmt.Errorf("level area not found: %w", err)
	}

	if selected.Level <= 1 {
		if parentID != nil {
			return errParentNotAllowed
		}
		return nil
	}

	if parentID == nil {
		return errParentRequired
	}
	if currentID != nil && *parentID == *currentID {
		return errInvalidParentLevel
	}

	var parent struct {
		ID              int `gorm:"column:id"`
		ParentAreaLevel int `gorm:"column:parent_area_level"`
	}

	err = tx.WithContext(ctx).
		Table("master_areas AS ma").
		Select("ma.id, la.level AS parent_area_level").
		Joins("JOIN level_areas AS la ON la.id::text = ma.level_area_id AND la.deleted_at IS NULL").
		Where("ma.id = ? AND ma.deleted_at IS NULL", *parentID).
		Take(&parent).Error
	if err != nil {
		return fmt.Errorf("parent master area not found: %w", err)
	}

	if parent.ParentAreaLevel != selected.Level-1 {
		return errInvalidParentLevel
	}

	return nil
}
