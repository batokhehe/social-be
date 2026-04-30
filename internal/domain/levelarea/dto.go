package levelarea

type CreateRequest struct {
	Level       int    `json:"level" validate:"required,min=1"`
	Name        string `json:"name" validate:"required,max=20"`
	Description string `json:"description" validate:"required,max=50"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	Level       int    `json:"level" validate:"required,min=1"`
	Name        string `json:"name" validate:"required,max=20"`
	Description string `json:"description" validate:"required,max=50"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}
