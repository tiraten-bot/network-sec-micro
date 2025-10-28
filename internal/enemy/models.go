package enemy

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EnemyType represents the type of enemy
type EnemyType string

const (
	EnemyTypeGoblin   EnemyType = "goblin"   // Steals coins
	EnemyTypePirate   EnemyType = "pirate"   // Steals weapons
	EnemyTypeDragon   EnemyType = "dragon"   // Boss enemy
	EnemyTypeSkeleton EnemyType = "skeleton" // Regular enemy
)

// Enemy represents an enemy in the system
type Enemy struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Type        EnemyType          `bson:"type" json:"type"`
	Level       int                `bson:"level" json:"level"`
	Health      int                `bson:"health" json:"health"`
	AttackPower int                `bson:"attack_power" json:"attack_power"`
	CreatedBy   string             `bson:"created_by" json:"created_by"` // Dark emperor/king username
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (Enemy) CollectionName() string {
	return "enemies"
}

// CanBeCreatedBy checks if a role can create enemies
func (et EnemyType) CanBeCreatedBy(role string) bool {
	// Only dark emperor and dark king can create enemies
	return role == "dark_emperor" || role == "dark_king"
}
