package masterdonationcategory

type CreateRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Status string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	Name   string `json:"name" validate:"required,max=100"`
	Status string `json:"status" validate:"omitempty,oneof=active inactive"`
}
