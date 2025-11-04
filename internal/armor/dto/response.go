package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArmorResponse represents an armor in responses
type ArmorResponse struct {
	ID           primitive.ObjectID `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Type         string             `json:"type"`
	Defense      int                `json:"defense"`
	HPBonus      int                `json:"hp_bonus"`
	Price        int                `json:"price"`
	CreatedBy    string             `json:"created_by"`
	OwnedBy      []string           `json:"owned_by"`
	Durability   int                `json:"durability"`
	MaxDurability int               `json:"max_durability"`
	IsBroken     bool               `json:"is_broken"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ArmorsListResponse represents a list of armors
type ArmorsListResponse struct {
	Armors []ArmorResponse `json:"armors"`
	Count  int            `json:"count"`
}

