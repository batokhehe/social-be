package user

import "time"

type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	VISID        string `json:"vis_id"`
	Role         int    `json:"role"`
	Status       int    `json:"status"`
	ProfilePhoto string `json:"profile_photo,omitempty"`

	// Latest successful login only (admin visibility); nil until the first login.
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP        *string    `json:"last_login_ip,omitempty"`
	LastLoginUserAgent *string    `json:"last_login_user_agent,omitempty"`
}
