package auth

type RegisterRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	ProfilePhoto string `json:"profile_photo,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginMeta is transport metadata captured at the HTTP layer for last-login
// tracking. It is NOT part of the login request/response contract.
type LoginMeta struct {
	IP        string
	UserAgent string
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	User         interface{} `json:"user,omitempty"`
	Data         interface{} `json:"data,omitempty"` // Volunteer data if role_id=2
}
