package masteractivity

type CreateRequest struct {
	CategoryActivityID int    `json:"category_activity_id" validate:"required,min=1"`
	Name               string `json:"name" validate:"required,max=50"`
	Target             int    `json:"target" validate:"required,min=0"`
	Description        string `json:"description" validate:"omitempty,max=200"`
	Status             string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	CategoryActivityID int    `json:"category_activity_id" validate:"required,min=1"`
	Name               string `json:"name" validate:"required,max=50"`
	Target             int    `json:"target" validate:"required,min=0"`
	Description        string `json:"description" validate:"omitempty,max=200"`
	Status             string `json:"status" validate:"omitempty,oneof=active inactive"`
}
