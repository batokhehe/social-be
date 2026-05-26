package email

import "time"

type SendRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type SendResponse struct {
	To          string    `json:"to"`
	Status      string    `json:"status"`
	Dummy       bool      `json:"dummy"`
	Message     string    `json:"message"`
	SimulatedAt time.Time `json:"simulated_at"`
}
