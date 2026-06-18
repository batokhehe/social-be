package donation

import "time"

// ImportError is a persisted per-row failure for an import batch. Column name
// `donator_id` follows the audit module's specified schema; it holds the raw
// "ID Donatur" value from the spreadsheet (which may not resolve to a donor).
type ImportError struct {
	ID           int       `json:"id" gorm:"column:id;primaryKey"`
	BatchID      string    `json:"batch_id" gorm:"column:batch_id"`
	RowNumber    int       `json:"row_number" gorm:"column:row_number"`
	DonaturID    string    `json:"donator_id" gorm:"column:donator_id"`
	CategoryName string    `json:"category_name" gorm:"column:category_name"`
	ErrorMessage string    `json:"error_message" gorm:"column:error_message"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at"`
}

func (ImportError) TableName() string { return "donation_import_errors" }
