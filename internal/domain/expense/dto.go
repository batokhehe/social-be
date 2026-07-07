package expense

// CreateRequest is the POST /expenses body. expense_no is auto-generated and is
// never accepted from the client.
type CreateRequest struct {
	ExpenseDate string  `json:"expense_date" validate:"required"`
	CategoryID  int     `json:"category_id" validate:"required,min=1"`
	VolunteerID int     `json:"volunteer_id" validate:"required,min=1"`
	Amount      float64 `json:"amount" validate:"gt=0"`
	Description string  `json:"description" validate:"omitempty,max=2000"`
	Status      string  `json:"status" validate:"required,oneof=draft paid cancelled"`
}

// UpdateRequest mirrors CreateRequest (expense_no stays immutable).
type UpdateRequest = CreateRequest
