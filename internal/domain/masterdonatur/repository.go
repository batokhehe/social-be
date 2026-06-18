package masterdonatur

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

// loadVolunteers resolves, for each donatur, the volunteer behind its
// id_vis_volunteer (parsed as volunteers.id) and that volunteer's master area,
// using one batched query per side. A donatur with an empty/non-numeric
// id_vis_volunteer, or one pointing to a missing/soft-deleted volunteer, keeps a
// nil Volunteer.
func (r *GormRepository) loadVolunteers(ctx context.Context, items []MasterDonatur) error {
	volIDSet := make(map[int]struct{})
	volIDs := make([]int, 0, len(items))
	for _, item := range items {
		if id, ok := parseVolunteerID(item.VisVolunteerID); ok {
			if _, seen := volIDSet[id]; !seen {
				volIDSet[id] = struct{}{}
				volIDs = append(volIDs, id)
			}
		}
	}
	if len(volIDs) == 0 {
		return nil
	}

	var volRows []struct {
		ID             int    `gorm:"column:id"`
		VISID          string `gorm:"column:vis_id"`
		IndonesianName string `gorm:"column:indonesian_name"`
		MasterAreaID   int    `gorm:"column:master_area_id"`
	}
	if err := r.DB.WithContext(ctx).
		Table("volunteers").
		Select("id", "vis_id", "indonesian_name", "master_area_id").
		Where("id IN ? AND deleted_at IS NULL", volIDs).
		Find(&volRows).Error; err != nil {
		r.Logger.WithError(err).Error("loadVolunteers MasterDonatur failed")
		return fmt.Errorf("failed load donatur volunteers: %w", err)
	}

	areaIDSet := make(map[int]struct{})
	areaIDs := make([]int, 0, len(volRows))
	for _, vr := range volRows {
		if vr.MasterAreaID > 0 {
			if _, ok := areaIDSet[vr.MasterAreaID]; !ok {
				areaIDSet[vr.MasterAreaID] = struct{}{}
				areaIDs = append(areaIDs, vr.MasterAreaID)
			}
		}
	}

	byAreaID := make(map[int]MasterAreaInfo)
	if len(areaIDs) > 0 {
		var areaRows []struct {
			ID       int    `gorm:"column:id"`
			Name     string `gorm:"column:name"`
			Location string `gorm:"column:location"`
		}
		if err := r.DB.WithContext(ctx).
			Table("master_areas").
			Select("id", "name", "location").
			Where("id IN ? AND deleted_at IS NULL", areaIDs).
			Find(&areaRows).Error; err != nil {
			r.Logger.WithError(err).Error("loadVolunteers (areas) MasterDonatur failed")
			return fmt.Errorf("failed load donatur volunteer areas: %w", err)
		}
		for _, ar := range areaRows {
			byAreaID[ar.ID] = MasterAreaInfo{ID: ar.ID, Name: ar.Name, Location: ar.Location}
		}
	}

	byVolID := make(map[int]DonaturVolunteer, len(volRows))
	for _, vr := range volRows {
		vol := DonaturVolunteer{
			ID:             vr.ID,
			VISID:          vr.VISID,
			IndonesianName: vr.IndonesianName,
			MasterAreaID:   vr.MasterAreaID,
		}
		if area, ok := byAreaID[vr.MasterAreaID]; ok {
			a := area
			vol.MasterArea = &a
		}
		byVolID[vr.ID] = vol
	}

	for i := range items {
		if id, ok := parseVolunteerID(items[i].VisVolunteerID); ok {
			if vol, found := byVolID[id]; found {
				v := vol
				items[i].Volunteer = &v
			}
		}
	}
	return nil
}

// loadDonaturGroups resolves each donatur's group (id_group_donatur ->
// master_donatur_groups.id) in one batched query. Donaturs with a NULL group or
// pointing to a missing/soft-deleted group keep a nil DonaturGroup.
func (r *GormRepository) loadDonaturGroups(ctx context.Context, items []MasterDonatur) error {
	idSet := make(map[int]struct{})
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if item.DonaturGroupID != nil && *item.DonaturGroupID > 0 {
			if _, ok := idSet[*item.DonaturGroupID]; !ok {
				idSet[*item.DonaturGroupID] = struct{}{}
				ids = append(ids, *item.DonaturGroupID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var rows []struct {
		ID      int    `gorm:"column:id"`
		GroupID string `gorm:"column:id_group_donatur"`
		Name    string `gorm:"column:name"`
		Status  string `gorm:"column:status"`
	}
	if err := r.DB.WithContext(ctx).
		Table("master_donatur_groups").
		Select("id", "id_group_donatur", "name", "status").
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("loadDonaturGroups MasterDonatur failed")
		return fmt.Errorf("failed load donatur groups: %w", err)
	}

	byID := make(map[int]DonaturGroupInfo, len(rows))
	for _, row := range rows {
		byID[row.ID] = DonaturGroupInfo{ID: row.ID, GroupID: row.GroupID, Name: row.Name, Status: row.Status}
	}

	for i := range items {
		if items[i].DonaturGroupID != nil {
			if g, ok := byID[*items[i].DonaturGroupID]; ok {
				ref := g
				items[i].DonaturGroup = &ref
			}
		}
	}
	return nil
}

// parseVolunteerID parses id_vis_volunteer into a volunteers PK. Returns
// (0, false) when empty or non-numeric (donatur without a volunteer).
func parseVolunteerID(visVolunteerID string) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(visVolunteerID))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
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

	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, err
	}
	if err := r.loadDonaturGroups(ctx, items); err != nil {
		return nil, err
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

	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, 0, err
	}
	if err := r.loadDonaturGroups(ctx, items); err != nil {
		return nil, 0, err
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

	items := []MasterDonatur{toEntity(item)}
	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, err
	}
	if err := r.loadDonaturGroups(ctx, items); err != nil {
		return nil, err
	}

	return &items[0], nil
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
