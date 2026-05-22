package speak

type CreateRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	PicUserID   int    `json:"pic" validate:"required"`
	CategoryID  *int   `json:"category_id,omitempty" validate:"omitempty,gt=0"`
	IsAnonymous bool   `json:"is_anonymous"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type UpdateRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	PicUserID   int    `json:"pic" validate:"required"`
	CategoryID  *int   `json:"category_id,omitempty" validate:"omitempty,gt=0"`
	IsAnonymous bool   `json:"is_anonymous"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type ActionRequest struct {
	Type string `json:"type" validate:"required,oneof=progress finish"`
	Note string `json:"note,omitempty" validate:"omitempty,max=1000"`
}
