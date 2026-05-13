package categoryactivity

type CreateRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description" validate:"omitempty,max=200"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description" validate:"omitempty,max=200"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}
