package weapon

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WeaponType represents the type of weapon
type WeaponType string

const (
	WeaponTypeCommon    WeaponType = "common"    // Basic weapons - Knight, Archer, Mage
	WeaponTypeRare      WeaponType = "rare"      // Rare weapons - Light King, Light Emperor
	WeaponTypeLegendary WeaponType = "legendary" // Legendary weapons - Only Emperor can buy
)

// Weapon represents a weapon in the system
type Weapon struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Type        WeaponType         `bson:"type" json:"type"`
	Damage      int                `bson:"damage" json:"damage"`
	Price       int                `bson:"price" json:"price"`
	CreatedBy   string             `bson:"created_by" json:"created_by"` // warrior username
	OwnedBy     []string           `bson:"owned_by" json:"owned_by"`     // list of warrior usernames
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (Weapon) CollectionName() string {
	return "weapons"
}

// CanBeCreatedBy checks if a role can create this weapon type
func (wt WeaponType) CanBeCreatedBy(role string) bool {
	// Only light emperor and light king can create weapons
	if role != "light_emperor" && role != "light_king" {
		return false
	}

	// Legendary weapons cannot be created by anyone
	if wt == WeaponTypeLegendary {
		return false
	}

	return true
}

// CanBeBoughtBy checks if a role can buy this weapon
func (w *Weapon) CanBeBoughtBy(role string) bool {
	// Dark side cannot buy weapons
	if role == "dark_emperor" || role == "dark_king" {
		return false
	}

	switch w.Type {
	case WeaponTypeCommon:
		// Common weapons can be bought by knights, archers, mages, kings, emperors
		return true

	case WeaponTypeRare:
		// Rare weapons can be bought by light king and light emperor
		return role == "light_king" || role == "light_emperor"

	case WeaponTypeLegendary:
		// Legendary weapons can only be bought by emperors
		return role == "light_emperor"

	default:
		return false
	}
}
