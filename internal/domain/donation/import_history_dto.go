package donation

import "time"

// ImportHistoryItem is one row in the import-history list. It mirrors the audit
// log enriched with the uploader's name. (komisariat == "commissioner".)
type ImportHistoryItem struct {
	BatchID          string     `json:"batch_id"`
	Filename         string     `json:"filename"`
	UploadedBy       *int       `json:"uploaded_by"`
	UploadedByName   string     `json:"uploaded_by_name"`
	UploadedAt       time.Time  `json:"uploaded_at"`
	KomisariatID     string     `json:"komisariat_id"`
	KomisariatName   string     `json:"komisariat_name"`
	Period           string     `json:"period"`
	TotalRows        int        `json:"total_rows"`
	SuccessRows      int        `json:"success_rows"`
	FailedRows       int        `json:"failed_rows"`
	SkippedRows      int        `json:"skipped_rows"`
	Status           string     `json:"status"`
	RolledBackBy     *int       `json:"rolled_back_by,omitempty"`
	RolledBackByName string     `json:"rolled_back_by_name,omitempty"`
	RolledBackAt     *time.Time `json:"rolled_back_at,omitempty"`
}

// ImportSummary is the row-count breakdown for a batch.
type ImportSummary struct {
	TotalRows   int `json:"total_rows"`
	SuccessRows int `json:"success_rows"`
	FailedRows  int `json:"failed_rows"`
	SkippedRows int `json:"skipped_rows"`
}

// ImportDetailResponse is the single-batch detail view.
type ImportDetailResponse struct {
	BatchID        string        `json:"batch_id"`
	Filename       string        `json:"filename"`
	UploadedBy     *int          `json:"uploaded_by"`
	UploadedByName string        `json:"uploaded_by_name"`
	UploadedAt     time.Time     `json:"uploaded_at"`
	KomisariatID   string        `json:"komisariat_id"`
	KomisariatName string        `json:"komisariat_name"`
	Period         string        `json:"period"`
	Status         string        `json:"status"`
	Summary        ImportSummary `json:"summary"`
	Rollback       RollbackInfo  `json:"rollback"`
}

// RollbackInfo answers "who rolled back this batch and when".
type RollbackInfo struct {
	RolledBack       bool       `json:"rolled_back"`
	Status           string     `json:"status"`
	RolledBackBy     *int       `json:"rolled_back_by"`
	RolledBackByName string     `json:"rolled_back_by_name,omitempty"`
	RolledBackAt     *time.Time `json:"rolled_back_at"`
}

// ImportHistorySort is a validated sort directive for the history list.
type ImportHistorySort struct {
	Column string
	Order  string // "asc" | "desc"
}

// OrderClause renders the SQL ORDER BY fragment. Inputs are pre-validated
// against a whitelist, so this is safe to interpolate.
func (s ImportHistorySort) OrderClause() string {
	return s.Column + " " + s.Order
}
