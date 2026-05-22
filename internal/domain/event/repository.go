package event

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
	Create(ctx context.Context, req CreateRequest, startAt, endAt time.Time, actorID int) (*Event, error)
	GetPaginated(ctx context.Context, page pagination.Query) ([]Event, int64, error)
	GetActiveEvents(ctx context.Context, now time.Time, page pagination.Query) ([]Event, int64, error)
	GetByID(ctx context.Context, id int) (*Event, error)
	Update(ctx context.Context, id int, req UpdateRequest, startAt, endAt time.Time, actorID int) (*Event, error)
	SoftDelete(ctx context.Context, id int, actorID int) error
	AddAttachment(ctx context.Context, eventID int, filePath string, description string, actorID int) (*EventAttachment, error)
	GetEventAttachments(ctx context.Context, eventID int) ([]EventAttachment, error)
	Apply(ctx context.Context, eventID int, volunteerID int, actorID int) (*EventAttendance, error)
	CheckIn(ctx context.Context, eventID int, volunteerID int, photoPath *string, actorID int) (*EventAttendance, error)
	CheckOut(ctx context.Context, eventID int, volunteerID int, photoPath *string, actorID int) (*EventAttendance, error)
}

type GormRepository struct {
	DB     *gorm.DB
	Logger *logrus.Logger
}

func NewGormRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &GormRepository{DB: db, Logger: logger}
}

type eventModel struct {
	ID                 int        `gorm:"column:id"`
	Name               string     `gorm:"column:name"`
	StartAt            time.Time  `gorm:"column:start_at"`
	EndAt              time.Time  `gorm:"column:end_at"`
	CategoryActivityID int        `gorm:"column:category_activity_id"`
	ActivityID         int        `gorm:"column:activity_id"`
	PicUserID          int        `gorm:"column:pic_user_id"`
	Status             string     `gorm:"column:status"`
	CreatedBy          *int       `gorm:"column:created_by"`
	UpdatedBy          *int       `gorm:"column:updated_by"`
	DeletedBy          *int       `gorm:"column:deleted_by"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
}

type eventAttachmentModel struct {
	ID          int        `gorm:"column:id"`
	EventID     int        `gorm:"column:event_id"`
	FilePath    string     `gorm:"column:file_path"`
	Description string     `gorm:"column:description"`
	CreatedBy   *int       `gorm:"column:created_by"`
	UpdatedBy   *int       `gorm:"column:updated_by"`
	DeletedBy   *int       `gorm:"column:deleted_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

type eventAttendanceModel struct {
	ID            int        `gorm:"column:id"`
	EventID       int        `gorm:"column:event_id"`
	VolunteerID   int        `gorm:"column:volunteer_id"`
	Status        string     `gorm:"column:status"`
	CheckinAt     *time.Time `gorm:"column:checkin_at"`
	CheckoutAt    *time.Time `gorm:"column:checkout_at"`
	CheckinPhoto  *string    `gorm:"column:checkin_photo"`
	CheckoutPhoto *string    `gorm:"column:checkout_photo"`
	CreatedBy     *int       `gorm:"column:created_by"`
	UpdatedBy     *int       `gorm:"column:updated_by"`
	DeletedBy     *int       `gorm:"column:deleted_by"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

type userModel struct {
	ID    int    `gorm:"column:id"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
}

func (eventModel) TableName() string           { return "events" }
func (eventAttachmentModel) TableName() string { return "event_attachments" }
func (eventAttendanceModel) TableName() string { return "event_attendances" }
func (userModel) TableName() string            { return "users" }

func toEntity(item eventModel) Event {
	return Event{
		ID:                 item.ID,
		Name:               item.Name,
		StartAt:            item.StartAt,
		EndAt:              item.EndAt,
		CategoryActivityID: item.CategoryActivityID,
		ActivityID:         item.ActivityID,
		PicUserID:          item.PicUserID,
		Status:             item.Status,
		CreatedBy:          item.CreatedBy,
		UpdatedBy:          item.UpdatedBy,
		DeletedBy:          item.DeletedBy,
	}
}

func toAttachmentEntity(item eventAttachmentModel) EventAttachment {
	return EventAttachment{
		ID:          item.ID,
		EventID:     item.EventID,
		FilePath:    item.FilePath,
		Description: item.Description,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
}

func toAttendanceEntity(item eventAttendanceModel) EventAttendance {
	attendance := EventAttendance{
		ID:          item.ID,
		EventID:     item.EventID,
		VolunteerID: item.VolunteerID,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
	}
	if item.CheckinAt != nil {
		val := item.CheckinAt.Format(time.RFC3339)
		attendance.CheckinAt = &val
	}
	if item.CheckoutAt != nil {
		val := item.CheckoutAt.Format(time.RFC3339)
		attendance.CheckoutAt = &val
	}
	attendance.CheckinPhoto = item.CheckinPhoto
	attendance.CheckoutPhoto = item.CheckoutPhoto
	return attendance
}

func (r *GormRepository) Create(ctx context.Context, req CreateRequest, startAt, endAt time.Time, actorID int) (*Event, error) {
	item := eventModel{
		Name:               req.Name,
		StartAt:            startAt,
		EndAt:              endAt,
		CategoryActivityID: req.CategoryActivityID,
		ActivityID:         req.ActivityID,
		PicUserID:          req.PicUserID,
		Status:             defaultString(req.Status, "active"),
		CreatedBy:          &actorID,
		UpdatedBy:          &actorID,
	}
	if err := r.DB.WithContext(ctx).Create(&item).Error; err != nil {
		r.Logger.WithError(err).Error("Create Event failed")
		return nil, fmt.Errorf("failed create event: %w", err)
	}
	return r.GetByID(ctx, item.ID)
}

func (r *GormRepository) GetPaginated(ctx context.Context, page pagination.Query) ([]Event, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&eventModel{}).Where("deleted_at IS NULL")
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed count events: %w", err)
	}
	var rows []eventModel
	if err := baseQuery.Order("start_at DESC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed get paginated events: %w", err)
	}
	items := make([]Event, 0, len(rows))
	for _, row := range rows {
		item := toEntity(row)
		_ = r.loadPicUser(ctx, &item)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) GetActiveEvents(ctx context.Context, now time.Time, page pagination.Query) ([]Event, int64, error) {
	var total int64
	baseQuery := r.DB.WithContext(ctx).Model(&eventModel{}).
		Where("status = ? AND deleted_at IS NULL AND end_at >= ?", "active", now)
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed count active events: %w", err)
	}
	var rows []eventModel
	if err := baseQuery.Order("start_at ASC").Limit(page.Limit).Offset(page.Offset).Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("failed get active events: %w", err)
	}
	items := make([]Event, 0, len(rows))
	for _, row := range rows {
		item := toEntity(row)
		_ = r.loadPicUser(ctx, &item)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id int) (*Event, error) {
	var item eventModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Take(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed get event by id: %w", err)
	}
	out := toEntity(item)
	if err := r.loadPicUser(ctx, &out); err != nil {
		return nil, err
	}
	if err := r.loadAttachments(ctx, &out); err != nil {
		return nil, err
	}
	if err := r.loadAttendances(ctx, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *GormRepository) Update(ctx context.Context, id int, req UpdateRequest, startAt, endAt time.Time, actorID int) (*Event, error) {
	result := r.DB.WithContext(ctx).Model(&eventModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"name":                 req.Name,
			"start_at":             startAt,
			"end_at":               endAt,
			"category_activity_id": req.CategoryActivityID,
			"activity_id":          req.ActivityID,
			"pic_user_id":          req.PicUserID,
			"status":               defaultString(req.Status, "active"),
			"updated_by":           actorID,
			"updated_at":           time.Now(),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("failed update event: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id int, actorID int) error {
	tx := r.DB.WithContext(ctx).Model(&eventModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"status":     "inactive",
			"deleted_at": time.Now(),
			"deleted_by": actorID,
			"updated_by": actorID,
			"updated_at": time.Now(),
		})
	if tx.Error != nil {
		return fmt.Errorf("failed delete event: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) AddAttachment(ctx context.Context, eventID int, filePath string, description string, actorID int) (*EventAttachment, error) {
	attachment := eventAttachmentModel{
		EventID:     eventID,
		FilePath:    filePath,
		Description: description,
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}
	if err := r.DB.WithContext(ctx).Create(&attachment).Error; err != nil {
		return nil, fmt.Errorf("failed create event attachment: %w", err)
	}
	out := toAttachmentEntity(attachment)
	return &out, nil
}

func (r *GormRepository) GetEventAttachments(ctx context.Context, eventID int) ([]EventAttachment, error) {
	var rows []eventAttachmentModel
	if err := r.DB.WithContext(ctx).
		Where("event_id = ? AND deleted_at IS NULL", eventID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed get event attachments: %w", err)
	}
	items := make([]EventAttachment, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAttachmentEntity(row))
	}
	return items, nil
}

func (r *GormRepository) Apply(ctx context.Context, eventID int, volunteerID int, actorID int) (*EventAttendance, error) {
	var event eventModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", eventID).Take(&event).Error; err != nil {
		return nil, err
	}
	if event.Status != "active" || event.EndAt.Before(time.Now()) {
		return nil, fmt.Errorf("event is not open for registration")
	}
	var existing eventAttendanceModel
	if err := r.DB.WithContext(ctx).
		Where("event_id = ? AND volunteer_id = ? AND deleted_at IS NULL", eventID, volunteerID).
		Take(&existing).Error; err == nil {
		return nil, fmt.Errorf("volunteer already registered for this event")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	attendance := eventAttendanceModel{
		EventID:     eventID,
		VolunteerID: volunteerID,
		Status:      "applied",
		CreatedBy:   &actorID,
		UpdatedBy:   &actorID,
	}
	if err := r.DB.WithContext(ctx).Create(&attendance).Error; err != nil {
		return nil, fmt.Errorf("failed apply event attendance: %w", err)
	}
	out := toAttendanceEntity(attendance)
	return &out, nil
}

func (r *GormRepository) CheckIn(ctx context.Context, eventID int, volunteerID int, photoPath *string, actorID int) (*EventAttendance, error) {
	var event eventModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", eventID).Take(&event).Error; err != nil {
		return nil, err
	}
	if event.Status != "active" || event.EndAt.Before(time.Now()) {
		return nil, fmt.Errorf("event is not open for check-in")
	}
	var attendance eventAttendanceModel
	if err := r.DB.WithContext(ctx).
		Where("event_id = ? AND volunteer_id = ? AND deleted_at IS NULL", eventID, volunteerID).
		Take(&attendance).Error; err != nil {
		return nil, err
	}
	if attendance.Status == "checked_out" {
		return nil, fmt.Errorf("attendance already checked out")
	}
	if attendance.Status == "checked_in" {
		return nil, fmt.Errorf("already checked in")
	}
	attendance.CheckinAt = ptrTime(time.Now())
	attendance.CheckinPhoto = photoPath
	attendance.Status = "checked_in"
	attendance.UpdatedBy = &actorID
	attendance.UpdatedAt = time.Now()
	if err := r.DB.WithContext(ctx).Save(&attendance).Error; err != nil {
		return nil, fmt.Errorf("failed check in: %w", err)
	}
	out := toAttendanceEntity(attendance)
	return &out, nil
}

func (r *GormRepository) CheckOut(ctx context.Context, eventID int, volunteerID int, photoPath *string, actorID int) (*EventAttendance, error) {
	var event eventModel
	if err := r.DB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", eventID).Take(&event).Error; err != nil {
		return nil, err
	}
	if event.Status != "active" || event.EndAt.Before(time.Now()) {
		return nil, fmt.Errorf("event is not open for checkout")
	}
	var attendance eventAttendanceModel
	if err := r.DB.WithContext(ctx).
		Where("event_id = ? AND volunteer_id = ? AND deleted_at IS NULL", eventID, volunteerID).
		Take(&attendance).Error; err != nil {
		return nil, err
	}
	if attendance.Status != "checked_in" {
		return nil, fmt.Errorf("must check in before checkout")
	}
	attendance.CheckoutAt = ptrTime(time.Now())
	attendance.CheckoutPhoto = photoPath
	attendance.Status = "checked_out"
	attendance.UpdatedBy = &actorID
	attendance.UpdatedAt = time.Now()
	if err := r.DB.WithContext(ctx).Save(&attendance).Error; err != nil {
		return nil, fmt.Errorf("failed check out: %w", err)
	}
	out := toAttendanceEntity(attendance)
	return &out, nil
}

func (r *GormRepository) loadAttachments(ctx context.Context, event *Event) error {
	attachments, err := r.GetEventAttachments(ctx, event.ID)
	if err != nil {
		return err
	}
	event.Attachments = attachments
	return nil
}

func (r *GormRepository) loadAttendances(ctx context.Context, event *Event) error {
	var rows []eventAttendanceModel
	if err := r.DB.WithContext(ctx).
		Where("event_id = ? AND deleted_at IS NULL", event.ID).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("failed get event attendances: %w", err)
	}
	items := make([]EventAttendance, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAttendanceEntity(row))
	}
	event.Attendances = items
	return nil
}

func (r *GormRepository) loadPicUser(ctx context.Context, event *Event) error {
	if event.PicUserID == 0 {
		return nil
	}
	var user userModel
	if err := r.DB.WithContext(ctx).Where("id = ?", event.PicUserID).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed load pic user: %w", err)
	}
	event.PicUser = &EventUserSummary{ID: user.ID, Name: user.Name, Email: user.Email}
	return nil
}

func defaultString(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
