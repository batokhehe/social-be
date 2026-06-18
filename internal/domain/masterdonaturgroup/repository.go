package masterdonaturgroup

import (
	"context"
	"errors"
	"fmt"
	"social-be/internal/pkg/pagination"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonaturGroup, error)
	GetAll(ctx context.Context) ([]MasterDonaturGroup, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonaturGroup, int64, error)
	GetByID(ctx context.Context, id int) (*MasterDonaturGroup, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonaturGroup, error)
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

type masterDonaturGroupModel struct {
	ID          int        `gorm:"column:id"`
	GroupID     string     `gorm:"column:id_group_donatur"`
	Name        string     `gorm:"column:name"`
	VolunteerID string     `gorm:"column:volunteer_id"`
	PICPhone    string     `gorm:"column:pic_phone"`
	Status      string     `gorm:"column:status"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

func (masterDonaturGroupModel) TableName() string {
	return "master_donatur_groups"
}

func toEntity(item masterDonaturGroupModel) MasterDonaturGroup {
	return MasterDonaturGroup{
		ID:          item.ID,
		GroupID:     item.GroupID,
		Name:        item.Name,
		VolunteerID: item.VolunteerID,
		PICPhone:    item.PICPhone,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
}

// volunteerPICRow is the minimal volunteers projection used to fill PIC.
type volunteerPICRow struct {
	ID             int    `gorm:"column:id"`
	IndonesianName string `gorm:"column:indonesian_name"`
	VISID          string `gorm:"column:vis_id"`
	Phone          string `gorm:"column:phone"`
}

// volunteerPK parses a group's volunteer_id ("123") into the volunteers PK.
// Returns (0, false) when it is empty or non-numeric.
func volunteerPK(volunteerID string) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(volunteerID))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// loadPICs resolves the volunteer behind each group's volunteer_id in a single
// query and attaches it as PIC. Groups whose volunteer_id is empty/non-numeric
// or points to a missing/soft-deleted volunteer keep a nil PIC.
func (r *GormRepository) loadPICs(ctx context.Context, items []MasterDonaturGroup) error {
	idSet := make(map[int]struct{})
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if id, ok := volunteerPK(item.VolunteerID); ok {
			if _, seen := idSet[id]; !seen {
				idSet[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var rows []volunteerPICRow
	if err := r.DB.WithContext(ctx).
		Table("volunteers").
		Select("id", "indonesian_name", "vis_id", "phone").
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("loadPICs MasterDonaturGroup failed")
		return fmt.Errorf("failed load donatur group pic: %w", err)
	}

	byID := make(map[int]PIC, len(rows))
	for _, row := range rows {
		byID[row.ID] = PIC{
			ID:             row.ID,
			IndonesianName: row.IndonesianName,
			VISID:          row.VISID,
			Phone:          row.Phone,
		}
	}

	for i := range items {
		if id, ok := volunteerPK(items[i].VolunteerID); ok {
			if pic, found := byID[id]; found {
				p := pic
				items[i].PIC = &p
			}
		}
	}
	return nil
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*MasterDonaturGroup, error) {
	item := masterDonaturGroupModel{
		GroupID:     req.GroupID,
		Name:        req.Name,
		VolunteerID: req.VolunteerID,
		PICPhone:    req.PICPhone,
		Status:      defaultString(req.Status, "active"),
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&item)
		if result.Error != nil {
			return fmt.Errorf("failed insert master donatur group: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("insert failed: no rows affected")
		}
		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Create MasterDonaturGroup failed")
		return nil, err
	}

	out := toEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAll(ctx context.Context) ([]MasterDonaturGroup, error) {
	var rows []masterDonaturGroupModel

	err := r.DB.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("id ASC").
		Find(&rows).Error

	if err != nil {
		r.Logger.WithError(err).Error("GetAll MasterDonaturGroup failed")
		return nil, fmt.Errorf("failed get all master donatur group: %w", err)
	}

	items := make([]MasterDonaturGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	if err := r.loadPICs(ctx, items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]MasterDonaturGroup, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&masterDonaturGroupModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count MasterDonaturGroup failed")
		return nil, 0, fmt.Errorf("failed count master donatur group: %w", err)
	}

	var rows []masterDonaturGroupModel
	if err := baseQuery.
		Order("id ASC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated MasterDonaturGroup failed")
		return nil, 0, fmt.Errorf("failed get paginated master donatur group: %w", err)
	}

	items := make([]MasterDonaturGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	if err := r.loadPICs(ctx, items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*MasterDonaturGroup, error) {
	var item masterDonaturGroupModel

	err := r.DB.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Take(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get master donatur group by id: %w", err)
	}

	items := []MasterDonaturGroup{toEntity(item)}
	if err := r.loadPICs(ctx, items); err != nil {
		return nil, err
	}

	return &items[0], nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*MasterDonaturGroup, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing masterDonaturGroupModel
		if err := tx.
			Where("id = ? AND deleted_at IS NULL", id).
			Take(&existing).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"id_group_donatur": req.GroupID,
			"name":             req.Name,
			"volunteer_id":     req.VolunteerID,
			"pic_phone":        req.PICPhone,
			"status":           defaultString(req.Status, existing.Status),
			"updated_by":       actorID,
			"updated_at":       time.Now(),
		}

		result := tx.Model(&masterDonaturGroupModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed update master donatur group: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})

	if err != nil {
		r.Logger.WithError(err).Error("Update MasterDonaturGroup failed")
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	result := r.DB.WithContext(ctx).
		Model(&masterDonaturGroupModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"deleted_by": actorID,
			"deleted_at": time.Now(),
		})

	if result.Error != nil {
		r.Logger.WithError(result.Error).Error("SoftDelete MasterDonaturGroup failed")
		return fmt.Errorf("failed delete master donatur group: %w", result.Error)
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
