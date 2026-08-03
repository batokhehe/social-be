package expense

// CreateRequest is the POST /expenses body. expense_no is auto-generated and is
// never accepted from the client. An expense belongs to a Master Area
// (required); volunteer_id is the optional PIC.
//
// Deliberately NO cross-field validation between master_area_id and
// volunteer_id: the PIC may belong to any Master Area, so the two are validated
// independently and never compared.
type CreateRequest struct {
	ExpenseDate  string `json:"expense_date" validate:"required"`
	MasterAreaID int    `json:"master_area_id" validate:"required,min=1"`
	CategoryID   int    `json:"category_id" validate:"required,min=1"`
	// VolunteerID is the optional PIC. Validated for existence only -- its own
	// master area is irrelevant and is never matched against MasterAreaID.
	VolunteerID *int    `json:"volunteer_id" validate:"omitempty,min=1"`
	Amount      float64 `json:"amount" validate:"gt=0"`
	Description string  `json:"description" validate:"omitempty,max=2000"`
	Status      string  `json:"status" validate:"required,oneof=draft paid cancelled"`
}

// UpdateRequest mirrors CreateRequest (expense_no stays immutable).
type UpdateRequest = CreateRequest
