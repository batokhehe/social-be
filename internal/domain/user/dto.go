package user

type CreateRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	ProfilePhoto string `json:"profile_photo,omitempty"`
}
