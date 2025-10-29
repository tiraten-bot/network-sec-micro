package dto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// CreateDragonCommand represents command to create a dragon
type CreateDragonCommand struct {
	Name          string
	Type          string
	Level         int
	CreatedBy     string
	CreatedByRole string
}

// AttackDragonCommand represents command to attack a dragon
type AttackDragonCommand struct {
	DragonID         primitive.ObjectID
	AttackerUsername string
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetDragonQuery represents query to get a dragon
type GetDragonQuery struct {
	DragonID primitive.ObjectID
}

// GetDragonsByTypeQuery represents query to get dragons by type
type GetDragonsByTypeQuery struct {
	Type      string
	AliveOnly bool
}

// GetDragonsByCreatorQuery represents query to get dragons by creator
type GetDragonsByCreatorQuery struct {
	CreatorUsername string
	AliveOnly       bool
}
