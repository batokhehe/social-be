package user

const (
	RoleSuperAdmin = 0
	RoleAdmin      = 1
	RoleVolunteer  = 2
)

type CreateRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	Role         *int   `json:"role,omitempty"`
	ProfilePhoto string `json:"profile_photo,omitempty"`
}
