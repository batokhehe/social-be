package masterexpensecategory

import "time"

// MasterExpenseCategory is the detail/list entity. The table has no audit-actor
// columns (created_by/updated_by/deleted_by), only timestamps.
type MasterExpenseCategory struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SelectItem is the lean dropdown shape for GET /select.
type SelectItem struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
