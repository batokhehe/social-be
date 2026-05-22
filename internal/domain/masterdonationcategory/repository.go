package masterdonationcategory

import (
	"context"
	"errors"
	"fmt"
	"social-be/internal/pkg/pagination"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonationCategory, error)
	GetAll(ctx context.Context) ([]MasterDonationCategory, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonationCategory, int64, error)
	GetByID(ctx context.Context, id int) (*MasterDonationCategory, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonationCategory, error)
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

type masterDonationCategoryModel struct {
	ID        int        `gorm:"column:id"`
	Name      string     `gorm:"column:name"`
	Status    string     `gorm:"column:status"`
	CreatedBy *int       `gorm:"column:created_by"`
	UpdatedBy *int       `gorm:"column:updated_by"`
	DeletedBy *int       `gorm:"column:deleted_by"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (masterDonationCategoryModel) TableName() string {
	return "master_donation_categories"
}

func toEntity(item masterDonationCategoryModel) MasterDonationCategory {
	return MasterDonationCategory{
		ID:        item.ID,
		Name:      item.Name,
		Status:    item.Status,
		CreatedBy: item.CreatedBy,
		UpdatedBy: item.UpdatedBy,
		DeletedBy: item.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonationCategory, error) {
	item := masterDonationCategoryModel{
		Name:      req.Name,
		Status:    defaultString(req.Status, "active"),
		CreatedBy: &actorID,
		UpdatedBy: &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&item)
		if result.Error != nil {
			return fmt.Errorf("failed insert master donation category: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("insert failed: no rows affected")
		}
		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create MasterDonationCategory failed")
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]MasterDonationCategory, error) {
	var rows []masterDonationCategoryModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error

	if err != nil {
		r.Logger.WithError(err).Error("GetAll MasterDonationCategory failed")
		return nil, fmt.Errorf("failed get all master donation category: %w", err)
	}

	items := make([]MasterDonationCategory, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonationCategory, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&masterDonationCategoryModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count MasterDonationCategory failed")
		return nil, 0, fmt.Errorf("failed count master donation category: %w", err)
	}

	var rows []masterDonationCategoryModel
	if err := baseQuery.
		Order("id ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated MasterDonationCategory failed")
		return nil, 0, fmt.Errorf("failed get paginated master donation category: %w", err)
	}

	items := make([]MasterDonationCategory, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterDonationCategory, error) {
	var item masterDonationCategoryModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get master donation category by id: %w", err)
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonationCategory, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing masterDonationCategoryModel
		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			Take(&existing).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"name":       req.Name,
			"status":     defaultString(req.Status, existing.Status),
			"updated_by": actorID,
			"updated_at": time.Now(),
		}

		result := tx.Model(&masterDonationCategoryModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed update master donation category: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update MasterDonationCategory failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	result := r.DB.WithContext(ctx).
		Model(&masterDonationCategoryModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_by": actorID,
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		r.Logger.WithError(result.Error).Error("SoftDelete MasterDonationCategory failed")
		return fmt.Errorf("failed delete master donation category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
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
