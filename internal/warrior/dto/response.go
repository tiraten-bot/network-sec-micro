package dto

import "time"

// LoginResponse represents a login response
type LoginResponse struct {
	Token   string         `json:"token"`
	Warrior WarriorResponse `json:"warrior"`
}

// WarriorResponse represents a warrior in responses
type WarriorResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
    Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WarriorsListResponse represents a list of warriors
type WarriorsListResponse struct {
	Role     string            `json:"role"`
	Warriors []WarriorResponse `json:"warriors"`
	Count    int               `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
