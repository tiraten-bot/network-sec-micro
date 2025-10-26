package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WeaponResponse represents a weapon in responses
type WeaponResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        string             `json:"type"`
	Damage      int                `json:"damage"`
	Price       int                `json:"price"`
	CreatedBy   string             `json:"created_by"`
	OwnedBy     []string           `json:"owned_by"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WeaponsListResponse represents a list of weapons
type WeaponsListResponse struct {
	Weapons []WeaponResponse `json:"weapons"`
	Count   int              `json:"count"`
}
