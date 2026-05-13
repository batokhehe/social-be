package levelarea

type CreateRequest struct {
	BelowLevelAreaID *int   `json:"below_level_area_id,omitempty" validate:"omitempty,min=1"`
	Name             string `json:"name" validate:"required,max=20"`
	Description      string `json:"description" validate:"required,max=50"`
	Status           string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateRequest struct {
	BelowLevelAreaID *int   `json:"below_level_area_id,omitempty" validate:"omitempty,min=1"`
	Name             string `json:"name" validate:"required,max=20"`
	Description      string `json:"description" validate:"required,max=50"`
	Status           string `json:"status" validate:"omitempty,oneof=active inactive"`
}
