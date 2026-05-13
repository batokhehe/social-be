package masterarea

type CreateRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	LevelAreaID string `json:"level_area_id" validate:"required,max=20"`
	Description string `json:"description" validate:"required,max=200"`
	ParentID    *int   `json:"parent_id,omitempty" validate:"omitempty,min=1"`
	Location    string `json:"location" validate:"required,max=100"`
}

type UpdateRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	LevelAreaID string `json:"level_area_id" validate:"required,max=20"`
	Description string `json:"description" validate:"required,max=200"`
	ParentID    *int   `json:"parent_id,omitempty" validate:"omitempty,min=1"`
	Location    string `json:"location" validate:"required,max=100"`
}
