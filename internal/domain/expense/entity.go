package expense

import "time"

// Expense is the detail entity (GET /expenses/:id). Nested relations are
// resolved read-only in the repository.
//
// Ownership rule: an expense belongs to a Master Area, and ONLY MasterAreaID
// determines ownership. VolunteerID is an optional PIC (person in charge) and
// may be absent. The PIC may belong to ANY Master Area -- MasterAreaID and
// Volunteer.MasterAreaID are allowed to differ and are never compared.
type Expense struct {
	ID           int             `json:"id"`
	ExpenseNo    string          `json:"expense_no"`
	ExpenseDate  time.Time       `json:"expense_date"`
	MasterAreaID int             `json:"master_area_id"`
	CategoryID   int             `json:"category_id"`
	VolunteerID  *int            `json:"volunteer_id"`
	Amount       float64         `json:"amount"`
	Description  string          `json:"description"`
	Status       string          `json:"status"`
	MasterArea   *MasterAreaInfo `json:"master_area,omitempty"`
	Category     *CategoryInfo   `json:"category,omitempty"`
	Volunteer    *VolunteerInfo  `json:"volunteer,omitempty"`
	CreatedBy    *UserInfo       `json:"created_by,omitempty"`
	UpdatedBy    *UserInfo       `json:"updated_by,omitempty"`
	DeletedBy    *UserInfo       `json:"deleted_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// MasterAreaInfo is the owning area resolved from master_area_id.
type MasterAreaInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

// CategoryInfo is the expense category resolved from category_id.
type CategoryInfo struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// VolunteerInfo is the optional PIC volunteer (null when the expense has no PIC).
// MasterAreaID here is the PIC's OWN area and is informational only: it may
// legitimately differ from the expense's MasterAreaID (cross-area PIC).
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

// ExpenseListItem is the lean list row (GET /expenses). AreaName is the owner;
// VolunteerName is the optional PIC and is empty when the expense has none.
type ExpenseListItem struct {
	ID            int       `json:"id"`
	ExpenseNo     string    `json:"expense_no"`
	ExpenseDate   time.Time `json:"expense_date"`
	Description   string    `json:"description"`
	AreaName      string    `json:"area_name"`
	CategoryName  string    `json:"category_name"`
	VolunteerName string    `json:"volunteer_name"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
