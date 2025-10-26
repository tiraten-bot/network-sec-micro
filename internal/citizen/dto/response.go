package dto

import "time"

// LoginResponse represents a login response
type LoginResponse struct {
	Token   string          `json:"token"`
	Citizen CitizenResponse `json:"citizen"`
}

// CitizenResponse represents a citizen in responses
type CitizenResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
