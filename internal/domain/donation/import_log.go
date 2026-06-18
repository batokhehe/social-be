package donation

import "time"

// Import log status values.
const (
	StatusCompleted          = "completed"
	StatusCompletedWithError = "completed_with_errors"
	StatusFailed             = "failed"
	StatusDryRun             = "dry_run"
	StatusRolledBack         = "rolled_back"
)

// ImportLog is the audit record for one Excel import (one upload = one batch).
// It is both the GORM model and the API representation.
type ImportLog struct {
	ID             int       `json:"id" gorm:"column:id;primaryKey"`
	BatchID        string    `json:"batch_id" gorm:"column:batch_id"`
	Filename       string    `json:"filename" gorm:"column:filename"`
	KomisariatID   string    `json:"komisariat_id" gorm:"column:komisariat_id"`
	KomisariatName string    `json:"komisariat_name" gorm:"column:komisariat_name"`
	Period         string    `json:"period" gorm:"column:period"`
	UploadedBy     *int      `json:"uploaded_by" gorm:"column:uploaded_by"`
	UploadedAt     time.Time `json:"uploaded_at" gorm:"column:uploaded_at"`
	TotalRows      int        `json:"total_rows" gorm:"column:total_rows"`
	SuccessRows    int        `json:"success_rows" gorm:"column:success_rows"`
	FailedRows     int        `json:"failed_rows" gorm:"column:failed_rows"`
	SkippedRows    int        `json:"skipped_rows" gorm:"column:skipped_rows"`
	Status         string     `json:"status" gorm:"column:status"`
	RolledBackBy   *int       `json:"rolled_back_by,omitempty" gorm:"column:rolled_back_by"`
	RolledBackAt   *time.Time `json:"rolled_back_at,omitempty" gorm:"column:rolled_back_at"`
}

func (ImportLog) TableName() string { return "donation_import_logs" }
