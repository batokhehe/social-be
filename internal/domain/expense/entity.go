package expense

import "time"

// Expense is the detail entity (GET /expenses/:id). Nested relations are
// resolved read-only in the repository; area is intentionally NOT stored here --
// it is derived through Volunteer.MasterAreaID.
type Expense struct {
	ID          int            `json:"id"`
	ExpenseNo   string         `json:"expense_no"`
	ExpenseDate time.Time      `json:"expense_date"`
	CategoryID  int            `json:"category_id"`
	VolunteerID int            `json:"volunteer_id"`
	Amount      float64        `json:"amount"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Category    *CategoryInfo  `json:"category,omitempty"`
	Volunteer   *VolunteerInfo `json:"volunteer,omitempty"`
	CreatedBy   *UserInfo      `json:"created_by,omitempty"`
	UpdatedBy   *UserInfo      `json:"updated_by,omitempty"`
	DeletedBy   *UserInfo      `json:"deleted_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CategoryInfo is the expense category resolved from category_id.
type CategoryInfo struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// VolunteerInfo is the PIC volunteer. MasterAreaID is exposed so dashboards
// (Hu Ai / Xie Li) can resolve area through the volunteer, never through expense.
type VolunteerInfo struct {
	ID             int    `json:"id"`
	IndonesianName string `json:"indonesian_name"`
	MasterAreaID   int    `json:"master_area_id"`
}

// UserInfo is an audit actor (created/updated/deleted by).
type UserInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ExpenseListItem is the lean list row (GET /expenses).
type ExpenseListItem struct {
	ID            int       `json:"id"`
	ExpenseNo     string    `json:"expense_no"`
	ExpenseDate   time.Time `json:"expense_date"`
	CategoryName  string    `json:"category_name"`
	VolunteerName string    `json:"volunteer_name"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
