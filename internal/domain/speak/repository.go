package speak

import (
	"context"
	"errors"
	"fmt"
	"time"

	"social-be/internal/pkg/pagination"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, req CreateRequest, actorID int) (*Speak, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]Speak, int64, error)
	GetPaginatedAsReporter(ctx context.Context, page pagination.Query, actorID int) ([]Speak, int64, error)
	GetPaginatedAsRespondent(ctx context.Context, page pagination.Query, actorID int) ([]Speak, int64, error)
	GetByID(ctx context.Context, id int) (*Speak, error)
	Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Speak, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
	Action(ctx context.Context, id int, req ActionRequest, actorID int) (*Speak, error)
	AddAttachment(ctx context.Context, speakID int, filePath string, originalName string, actorID int, attachmentType int) (*SpeakAttachment, error)
	GetAttachments(ctx context.Context, speakID int, attachmentType int) ([]SpeakAttachment, error)
	SoftDeleteAttachment(ctx context.Context, speakID int, attachmentID int, actorID int) error
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, actorID int) (*Speak, error) {
	item := speakModel{
		Name:        req.Name,
		PicUserID:   req.PicUserID,
		CategoryID:  req.CategoryID,
		IsAnonymous: req.IsAnonymous,
		Description: req.Description,
		Status:      0,
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}

	if err := r.DB.WithContext(ctx).Create(&item).Error; err != nil {
		r.Logger.WithError(err).Error("Create Speak failed")
		return nil, fmt.Errorf("failed create speak: %w", err)
	}

	return r.GetByID(ctx, item.ID)
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]Speak, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&speakModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count Speak failed")
		return nil, 0, fmt.Errorf("failed count speaks: %w", err)
	}

	var rows []speakModel
	if err := baseQuery.Order("id DESC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginated Speak failed")
		return nil, 0, fmt.Errorf("failed get paginated speaks: %w", err)
	}

	items := make([]Speak, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *GormRepository) GetPaginatedAsReporter(ctx context.Context, page pagination.Query, actorID int) ([]Speak, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&speakModel{}).Where("deleted_at IS NULL AND created_by = ?", actorID)
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count Speak failed")
		return nil, 0, fmt.Errorf("failed count speaks: %w", err)
	}

	var rows []speakModel
	if err := baseQuery.Order("id DESC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginatedAsReporter Speak failed")
		return nil, 0, fmt.Errorf("failed get paginated speaks: %w", err)
	}

	items := make([]Speak, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *GormRepository) GetPaginatedAsRespondent(ctx context.Context, page pagination.Query, actorID int) ([]Speak, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&speakModel{}).Where("deleted_at IS NULL AND pic_user_id = ?", actorID)
	if err := baseQuery.Count(&total).Error; err != nil {
		r.Logger.WithError(err).Error("Count Speak failed")
		return nil, 0, fmt.Errorf("failed count speaks: %w", err)
	}

	var rows []speakModel
	if err := baseQuery.Order("id DESC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("GetPaginatedAsReporter Speak failed")
		return nil, 0, fmt.Errorf("failed get paginated speaks: %w", err)
	}

	items := make([]Speak, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}

	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Speak, error) {
	var item speakModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get speak by id: %w", err)
	}

	items := []Speak{toEntity(item)}
	if err := r.loadAttachments(ctx, &items[0], 0); err != nil {
		return nil, err
	}
	if err := r.loadVolunteers(ctx, items); err != nil {
		return nil, err
	}

	return &items[0], nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, actorID int) (*Speak, error) {
	result := r.DB.WithContext(ctx).Model(&speakModel{}).Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"name":         req.Name,
			"pic_user_id":  req.PicUserID,
			"category_id":  req.CategoryID,
			"is_anonymous": req.IsAnonymous,
			"description":  req.Description,
			"updated_by":   actorID,
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		r.Logger.WithError(result.Error).Error("Update Speak failed")
		return nil, fmt.Errorf("failed update speak: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&speakModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"deleted_at": time.Now(),
		"deleted_by": actorID,
		"updated_by": actorID,
		"updated_at": time.Now(),
	})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("SoftDelete Speak failed")
		return fmt.Errorf("failed soft delete speak: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRepository) Action(ctx context.Context, id int, req ActionRequest, actorID int) (*Speak, error) {
	updates := map[string]any{
		"updated_by": actorID,
		"updated_at": time.Now(),
	}

	if req.Type == "progress" {
		updates["status"] = 1
		updates["progress_at"] = time.Now()
		updates["progress_notes"] = nullIfEmpty(req.Note)
	} else {
		updates["status"] = 2
		updates["finish_at"] = time.Now()
		updates["finish_notes"] = nullIfEmpty(req.Note)
	}

	result := r.DB.WithContext(ctx).Model(&speakModel{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
	if result.Error != nil {
		r.Logger.WithError(result.Error).Error("Action Speak failed")
		return nil, fmt.Errorf("failed update speak action: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *GormRepository) AddAttachment(ctx context.Context, speakID int, filePath string, originalName string, actorID int, attachmentType int) (*SpeakAttachment, error) {
	item := speakAttachmentModel{
		SpeakID:      speakID,
		FilePath:     filePath,
		OriginalName: originalName,
		CreatedBy:    &actorID,
		UpdatedBy:    &actorID,
		Type:         attachmentType,
	}

	if err := r.DB.WithContext(ctx).Create(&item).Error; err != nil {
		r.Logger.WithError(err).Error("Create SpeakAttachment failed")
		return nil, fmt.Errorf("failed add speak attachment: %w", err)
	}

	out := toAttachmentEntity(item)
	return &out, nil
}

func (r *GormRepository) GetAttachments(ctx context.Context, speakID int, attachmentType int) ([]SpeakAttachment, error) {
	var rows []speakAttachmentModel
	if err := r.DB.WithContext(ctx).Where("speak_id = ? AND type = ? AND deleted_at IS NULL", speakID, attachmentType).Order("id ASC").Find(&rows).Error; err != nil {
		r.Logger.WithError(err).Error("Get SpeakAttachments failed")
		return nil, fmt.Errorf("failed get speak attachments: %w", err)
	}

	items := make([]SpeakAttachment, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAttachmentEntity(row))
	}

	return items, nil
}

func (r *GormRepository) SoftDeleteAttachment(ctx context.Context, speakID int, attachmentID int, actorID int) error {
	// Scope by speak_id as well so an attachment can only be deleted through its
	// owning speak.
	tx := r.DB.WithContext(ctx).Model(&speakAttachmentModel{}).
		Where("id = ? AND speak_id = ? AND deleted_at IS NULL", attachmentID, speakID).
		Updates(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": actorID,
			"updated_by": actorID,
			"updated_at": time.Now(),
		})
	if tx.Error != nil {
		r.Logger.WithError(tx.Error).Error("SoftDelete SpeakAttachment failed")
		return fmt.Errorf("failed soft delete speak attachment: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormRepository) loadAttachments(ctx context.Context, speak *Speak, attachmentType int) error {
	items, err := r.GetAttachments(ctx, speak.ID, attachmentType)
	if err != nil {
		return err
	}

	speak.Attachments = items
	return nil
}

// speakVolunteerRow is the minimal volunteers projection used to fill the PIC
// and reporter (created_by) volunteer references on a speak.
type speakVolunteerRow struct {
	ID             int    `gorm:"column:id"`
	UserID         int    `gorm:"column:user_id"`
	IndonesianName string `gorm:"column:indonesian_name"`
	VISID          string `gorm:"column:vis_id"`
	Phone          string `gorm:"column:phone"`
}

func (row speakVolunteerRow) toRef() SpeakVolunteer {
	return SpeakVolunteer{
		ID:             row.ID,
		UserID:         row.UserID,
		IndonesianName: row.IndonesianName,
		VISID:          row.VISID,
		Phone:          row.Phone,
	}
}

// loadVolunteers resolves, for every speak in items:
//   - PICVolunteer       from pic (pic_user_id) -> volunteers.id
//   - CreatedByVolunteer from created_by (users.id) -> volunteers.user_id
//
// Each side is loaded in a single batched query. Missing / soft-deleted
// volunteers simply leave the corresponding reference nil.
func (r *GormRepository) loadVolunteers(ctx context.Context, items []Speak) error {
	picIDSet := make(map[int]struct{})
	picIDs := make([]int, 0, len(items))
	userIDSet := make(map[int]struct{})
	userIDs := make([]int, 0, len(items))

	for _, item := range items {
		if item.PicUserID > 0 {
			if _, ok := picIDSet[item.PicUserID]; !ok {
				picIDSet[item.PicUserID] = struct{}{}
				picIDs = append(picIDs, item.PicUserID)
			}
		}
		if item.CreatedBy != nil && *item.CreatedBy > 0 {
			if _, ok := userIDSet[*item.CreatedBy]; !ok {
				userIDSet[*item.CreatedBy] = struct{}{}
				userIDs = append(userIDs, *item.CreatedBy)
			}
		}
	}

	byVolunteerID := make(map[int]SpeakVolunteer)
	if len(picIDs) > 0 {
		var rows []speakVolunteerRow
		if err := r.DB.WithContext(ctx).
			Table("volunteers").
			Select("id", "user_id", "indonesian_name", "vis_id", "phone").
			Where("id IN ? AND deleted_at IS NULL", picIDs).
			Find(&rows).Error; err != nil {
			r.Logger.WithError(err).Error("loadVolunteers (pic) failed")
			return fmt.Errorf("failed load speak pic volunteer: %w", err)
		}
		for _, row := range rows {
			byVolunteerID[row.ID] = row.toRef()
		}
	}

	byUserID := make(map[int]SpeakVolunteer)
	if len(userIDs) > 0 {
		var rows []speakVolunteerRow
		if err := r.DB.WithContext(ctx).
			Table("volunteers").
			Select("id", "user_id", "indonesian_name", "vis_id", "phone").
			Where("user_id IN ? AND deleted_at IS NULL", userIDs).
			Find(&rows).Error; err != nil {
			r.Logger.WithError(err).Error("loadVolunteers (created_by) failed")
			return fmt.Errorf("failed load speak reporter volunteer: %w", err)
		}
		for _, row := range rows {
			byUserID[row.UserID] = row.toRef()
		}
	}

	for i := range items {
		if v, ok := byVolunteerID[items[i].PicUserID]; ok {
			ref := v
			items[i].PICVolunteer = &ref
		}
		if items[i].CreatedBy != nil {
			if v, ok := byUserID[*items[i].CreatedBy]; ok {
				ref := v
				items[i].CreatedByVolunteer = &ref
			}
		}
	}
	return nil
}

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
