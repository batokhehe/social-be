package masterexpensecategory

// CreateRequest is the POST body. Active is a *bool so "required" distinguishes
// an omitted field from an explicit false (validator "required" fails on false
// for a plain bool).
type CreateRequest struct {
	Code   string `json:"code" validate:"required,max=30"`
	Name   string `json:"name" validate:"required,max=100"`
	Active *bool  `json:"active" validate:"required"`
}

// UpdateRequest mirrors CreateRequest.
type UpdateRequest = CreateRequest
