package event

import "time"

type Event struct {
	ID                 int               `json:"id"`
	Name               string            `json:"name"`
	StartAt            time.Time         `json:"start_at"`
	EndAt              time.Time         `json:"end_at"`
	CategoryActivityID int               `json:"category_activity_id"`
	ActivityID         int               `json:"activity_id"`
	PicUserID          int               `json:"pic_user_id"`
	PicUser            *EventUserSummary `json:"pic_user,omitempty"`
	Status             string            `json:"status"`
	Attachments        []EventAttachment `json:"attachments,omitempty"`
	Attendances        []EventAttendance `json:"attendances,omitempty"`
	CreatedBy          *int              `json:"created_by,omitempty"`
	UpdatedBy          *int              `json:"updated_by,omitempty"`
	DeletedBy          *int              `json:"deleted_by,omitempty"`
	IsApplied          *bool             `json:"is_applied"`
}

type EventAttachment struct {
	ID          int    `json:"id"`
	EventID     int    `json:"event_id"`
	FilePath    string `json:"file_path"`
	Description string `json:"description,omitempty"`
	CreatedBy   *int   `json:"created_by,omitempty"`
	UpdatedBy   *int   `json:"updated_by,omitempty"`
	DeletedBy   *int   `json:"deleted_by,omitempty"`
}

type EventAttendance struct {
	ID            int               `json:"id"`
	EventID       int               `json:"event_id"`
	VolunteerID   int               `json:"volunteer_id"`
	Status        string            `json:"status"`
	CheckinAt     *string           `json:"checkin_at,omitempty"`
	CheckoutAt    *string           `json:"checkout_at,omitempty"`
	CheckinPhoto  *string           `json:"checkin_photo,omitempty"`
	CheckoutPhoto *string           `json:"checkout_photo,omitempty"`
	CreatedBy     *int              `json:"created_by,omitempty"`
	UpdatedBy     *int              `json:"updated_by,omitempty"`
	DeletedBy     *int              `json:"deleted_by,omitempty"`
	Volunteer     *VolunteerSummary `json:"volunteer,omitempty"`
}

type EventUserSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
