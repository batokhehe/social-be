package speak

import "time"

type Speak struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	PicUserID     int               `json:"pic"`
	CategoryID    *int              `json:"category_id,omitempty"`
	IsAnonymous   bool              `json:"is_anonymous"`
	Description   string            `json:"description,omitempty"`
	Status        int               `json:"status"`
	ProgressAt    *string           `json:"progress_at,omitempty"`
	ProgressNotes *string           `json:"progress_notes,omitempty"`
	FinishAt      *string           `json:"finish_at,omitempty"`
	FinishNotes   *string           `json:"finish_notes,omitempty"`
	CreatedBy     *int              `json:"created_by,omitempty"`
	UpdatedBy     *int              `json:"updated_by,omitempty"`
	DeletedBy     *int              `json:"deleted_by,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     *time.Time        `json:"deleted_at,omitempty"`
	Attachments   []SpeakAttachment `json:"attachments,omitempty"`
	// PICVolunteer is the volunteer in charge, resolved from pic (pic_user_id)
	// against volunteers.id. CreatedByVolunteer is the reporter, resolved from
	// created_by (users.id) against volunteers.user_id. Both are null if no
	// matching volunteer exists.
	PICVolunteer       *SpeakVolunteer `json:"pic_volunteer"`
	CreatedByVolunteer *SpeakVolunteer `json:"created_by_volunteer"`
}

// SpeakVolunteer is the volunteer reference embedded in a speak (PIC / reporter).
type SpeakVolunteer struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	IndonesianName string `json:"indonesian_name"`
	VISID          string `json:"vis_id"`
	Phone          string `json:"phone"`
}

// Speak attachment owner types.
const (
	AttachmentTypeReporter   = 0
	AttachmentTypeRespondent = 1
)

type SpeakAttachment struct {
	ID           int        `json:"id"`
	SpeakID      int        `json:"speak_id"`
	FilePath     string     `json:"file_path"`
	OriginalName string     `json:"original_name"`
	Type         int        `json:"type"` // 0 = reporter, 1 = respondent
	CreatedBy    *int       `json:"created_by,omitempty"`
	UpdatedBy    *int       `json:"updated_by,omitempty"`
	DeletedBy    *int       `json:"deleted_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type speakModel struct {
	ID            int        `gorm:"column:id"`
	Name          string     `gorm:"column:name"`
	PicUserID     int        `gorm:"column:pic_user_id"`
	CategoryID    *int       `gorm:"column:category_id"`
	IsAnonymous   bool       `gorm:"column:is_anonymous"`
	Description   string     `gorm:"column:description"`
	Status        int        `gorm:"column:status"`
	ProgressAt    *time.Time `gorm:"column:progress_at"`
	ProgressNotes *string    `gorm:"column:progress_notes"`
	FinishAt      *time.Time `gorm:"column:finish_at"`
	FinishNotes   *string    `gorm:"column:finish_notes"`
	CreatedBy     *int       `gorm:"column:created_by"`
	UpdatedBy     *int       `gorm:"column:updated_by"`
	DeletedBy     *int       `gorm:"column:deleted_by"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

type speakAttachmentModel struct {
	ID           int        `gorm:"column:id"`
	SpeakID      int        `gorm:"column:speak_id"`
	FilePath     string     `gorm:"column:file_path"`
	OriginalName string     `gorm:"column:original_name"`
	CreatedBy    *int       `gorm:"column:created_by"`
	UpdatedBy    *int       `gorm:"column:updated_by"`
	DeletedBy    *int       `gorm:"column:deleted_by"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	Type         int        `gorm:"column:type"`
}

func (speakModel) TableName() string {
	return "speaks"
}

func (speakAttachmentModel) TableName() string {
	return "speak_attachments"
}

func toEntity(item speakModel) Speak {
	out := Speak{
		ID:          item.ID,
		Name:        item.Name,
		PicUserID:   item.PicUserID,
		CategoryID:  item.CategoryID,
		IsAnonymous: item.IsAnonymous,
		Description: item.Description,
		Status:      item.Status,
		CreatedBy:   item.CreatedBy,
		UpdatedBy:   item.UpdatedBy,
		DeletedBy:   item.DeletedBy,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		DeletedAt:   item.DeletedAt,
	}

	if item.ProgressAt != nil {
		val := item.ProgressAt.Format(time.RFC3339)
		out.ProgressAt = &val
	}
	if item.ProgressNotes != nil {
		out.ProgressNotes = item.ProgressNotes
	}
	if item.FinishAt != nil {
		val := item.FinishAt.Format(time.RFC3339)
		out.FinishAt = &val
	}
	if item.FinishNotes != nil {
		out.FinishNotes = item.FinishNotes
	}

	return out
}

func toAttachmentEntity(item speakAttachmentModel) SpeakAttachment {
	return SpeakAttachment{
		ID:           item.ID,
		SpeakID:      item.SpeakID,
		FilePath:     item.FilePath,
		OriginalName: item.OriginalName,
		Type:         item.Type,
		CreatedBy:    item.CreatedBy,
		UpdatedBy:    item.UpdatedBy,
		DeletedBy:    item.DeletedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		DeletedAt:    item.DeletedAt,
	}
}
