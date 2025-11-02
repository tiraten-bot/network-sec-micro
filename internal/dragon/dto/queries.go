package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Dragon represents a dragon (imported from models)
type Dragon struct {
	ID                        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                      string             `bson:"name" json:"name"`
	Type                      string             `bson:"type" json:"type"`
	Level                     int                `bson:"level" json:"level"`
	Health                    int                `bson:"health" json:"health"`
	MaxHealth                 int                `bson:"max_health" json:"max_health"`
	AttackPower               int                `bson:"attack_power" json:"attack_power"`
	Defense                   int                `bson:"defense" json:"defense"`
	CreatedBy                 string             `bson:"created_by" json:"created_by"`
	IsAlive                   bool               `bson:"is_alive" json:"is_alive"`
	KilledBy                  string             `bson:"killed_by,omitempty" json:"killed_by,omitempty"`
	KilledAt                  *string            `bson:"killed_at,omitempty" json:"killed_at,omitempty"`
	RevivalCount              int                `bson:"revival_count" json:"revival_count"`
	AwaitingCrisisIntervention bool               `bson:"awaiting_crisis_intervention" json:"awaiting_crisis_intervention"`
	CreatedAt                 string             `bson:"created_at" json:"created_at"`
	UpdatedAt                 string             `bson:"updated_at" json:"updated_at"`
}

// ==================== HTTP REQUEST/RESPONSE DTOs ====================

// CreateDragonRequest represents HTTP request to create a dragon
type CreateDragonRequest struct {
	Name  string `json:"name" binding:"required"`
	Type  string `json:"type" binding:"required,oneof=fire ice lightning shadow"`
	Level int    `json:"level" binding:"required,min=1,max=100"`
}

// CreateDragonResponse represents HTTP response for dragon creation
type CreateDragonResponse struct {
	Success bool   `json:"success"`
	Dragon  *Dragon `json:"dragon"`
	Message string `json:"message"`
}

// AttackDragonRequest represents HTTP request to attack a dragon
type AttackDragonRequest struct {
	DragonID primitive.ObjectID `json:"dragon_id" binding:"required"`
}

// AttackDragonResponse represents HTTP response for dragon attack
type AttackDragonResponse struct {
	Success bool   `json:"success"`
	Dragon  *Dragon `json:"dragon"`
	Message string `json:"message"`
}

// GetDragonResponse represents HTTP response for getting a dragon
type GetDragonResponse struct {
	Success bool   `json:"success"`
	Dragon  *Dragon `json:"dragon"`
}

// GetDragonsByTypeResponse represents HTTP response for getting dragons by type
type GetDragonsByTypeResponse struct {
	Success bool     `json:"success"`
	Dragons []Dragon `json:"dragons"`
	Count   int      `json:"count"`
}

// GetDragonsByCreatorResponse represents HTTP response for getting dragons by creator
type GetDragonsByCreatorResponse struct {
	Success bool     `json:"success"`
	Dragons []Dragon `json:"dragons"`
	Count   int      `json:"count"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
