package masterdonatur

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
	Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonatur, error)
	GetAll(ctx context.Context) ([]MasterDonatur, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonatur, int64, error)
	GetByID(ctx context.Context, id int) (*MasterDonatur, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonatur, error)
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

type masterDonaturModel struct {
	ID             int        `gorm:"column:id"`
	DonaturID      string     `gorm:"column:id_donatur"`
	Phone          string     `gorm:"column:telepon"`
	TzuChiAppID    string     `gorm:"column:id_tzu_chi_app"`
	VisVolunteerID string     `gorm:"column:id_vis_volunteer"`
	DonaturGroupID *int       `gorm:"column:id_group_donatur"`
	Name           string     `gorm:"column:name"`
	Status         string     `gorm:"column:status"`
	CreatedBy      *int       `gorm:"column:created_by"`
	UpdatedBy      *int       `gorm:"column:updated_by"`
	DeletedBy      *int       `gorm:"column:deleted_by"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

func (masterDonaturModel) TableName() string {
	return "master_donaturs"
}

func toEntity(item masterDonaturModel) MasterDonatur {
	return MasterDonatur{
		ID:             item.ID,
		DonaturID:      item.DonaturID,
		Phone:          item.Phone,
		TzuChiAppID:    item.TzuChiAppID,
		VisVolunteerID: item.VisVolunteerID,
		DonaturGroupID: item.DonaturGroupID,
		Name:           item.Name,
		Status:         item.Status,
		CreatedBy:      item.CreatedBy,
		UpdatedBy:      item.UpdatedBy,
		DeletedBy:      item.DeletedBy,
	}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonatur, error) {
	item := masterDonaturModel{
		DonaturID:      req.DonaturID,
		Phone:          req.Phone,
		TzuChiAppID:    req.TzuChiAppID,
		VisVolunteerID: req.VisVolunteerID,
		DonaturGroupID: req.DonaturGroupID,
		Name:           req.Name,
		Status:         defaultString(req.Status, "active"),
		CreatedBy:      &actorID,
		UpdatedBy:      &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&item)
		if result.Error != nil {
			return fmt.Errorf("failed insert master donatur: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("insert failed: no rows affected")
		}
		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create MasterDonatur failed")
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]MasterDonatur, error) {
	var rows []masterDonaturModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error

	if err != nil {
		r.Logger.WithError(err).Error("GetAll MasterDonatur failed")
		return nil, fmt.Errorf("failed get all master donatur: %w", err)
	}

	items := make([]MasterDonatur, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonatur, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&masterDonaturModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count MasterDonatur failed")
		return nil, 0, fmt.Errorf("failed count master donatur: %w", err)
	}

	var rows []masterDonaturModel
	if err := baseQuery.
		Order("id ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated MasterDonatur failed")
		return nil, 0, fmt.Errorf("failed get paginated master donatur: %w", err)
	}

	items := make([]MasterDonatur, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterDonatur, error) {
	var item masterDonaturModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get master donatur by id: %w", err)
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonatur, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing masterDonaturModel
		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			Take(&existing).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"id_donatur":       req.DonaturID,
			"telepon":          req.Phone,
			"id_tzu_chi_app":   req.TzuChiAppID,
			"id_vis_volunteer":   req.VisVolunteerID,
			"id_group_donatur": req.DonaturGroupID,
			"name":             req.Name,
			"status":           defaultString(req.Status, existing.Status),
			"updated_by":       actorID,
			"updated_at":       time.Now(),
		}

		result := tx.Model(&masterDonaturModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed update master donatur: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update MasterDonatur failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	result := r.DB.WithContext(ctx).
		Model(&masterDonaturModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_by": actorID,
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		r.Logger.WithError(result.Error).Error("SoftDelete MasterDonatur failed")
		return fmt.Errorf("failed delete master donatur: %w", result.Error)
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
