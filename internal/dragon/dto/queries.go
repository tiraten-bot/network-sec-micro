package dragon

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
