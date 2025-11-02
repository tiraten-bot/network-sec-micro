package dragon

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DragonType represents the type of dragon
type DragonType string

const (
	DragonTypeFire    DragonType = "fire"    // Fire dragon
	DragonTypeIce     DragonType = "ice"     // Ice dragon
	DragonTypeLightning DragonType = "lightning" // Lightning dragon
	DragonTypeShadow  DragonType = "shadow"  // Shadow dragon
)

// Dragon represents a dragon in the system
type Dragon struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Type        DragonType         `bson:"type" json:"type"`
	Level       int                `bson:"level" json:"level"`
	Health      int                `bson:"health" json:"health"`
	MaxHealth   int                `bson:"max_health" json:"max_health"`
	AttackPower int                `bson:"attack_power" json:"attack_power"`
	Defense     int                `bson:"defense" json:"defense"`
	CreatedBy   string             `bson:"created_by" json:"created_by"` // Dark emperor username
	IsAlive     bool               `bson:"is_alive" json:"is_alive"`
	KilledBy    string             `bson:"killed_by,omitempty" json:"killed_by,omitempty"` // Light king/emperor username
	KilledAt    *time.Time         `bson:"killed_at,omitempty" json:"killed_at,omitempty"`
	RevivalCount int               `bson:"revival_count" json:"revival_count"` // Number of times dragon has been revived (max 3)
	AwaitingCrisisIntervention bool `bson:"awaiting_crisis_intervention" json:"awaiting_crisis_intervention"` // True when revival_count == 2 (before 3rd revival)
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (Dragon) CollectionName() string {
	return "dragons"
}

// CanBeCreatedBy checks if a role can create dragons
func (dt DragonType) CanBeCreatedBy(role string) bool {
	// Only dark emperor can create dragons
	return role == "dark_emperor"
}

// CanBeKilledBy checks if a role can kill dragons
func (d *Dragon) CanBeKilledBy(role string) bool {
	// Only light king (3x) or light emperor (1x) can kill dragons
	return role == "light_king" || role == "light_emperor"
}

// IsDead checks if dragon is dead
func (d *Dragon) IsDead() bool {
	return !d.IsAlive
}

// CanRevive checks if dragon can still be revived (max 3 revivals)
func (d *Dragon) CanRevive() bool {
	return d.RevivalCount < 3
}

// NeedsCrisisIntervention checks if dragon needs dark emperor intervention before 3rd revival
func (d *Dragon) NeedsCrisisIntervention() bool {
	return d.RevivalCount == 2 && !d.IsAlive
}

// TakeDamage reduces dragon's health
func (d *Dragon) TakeDamage(damage int) {
	if d.Health > damage {
		d.Health -= damage
	} else {
		d.Health = 0
		d.IsAlive = false
	}
}
