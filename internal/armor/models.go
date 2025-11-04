package armor

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArmorType represents the type of armor
type ArmorType string

const (
	ArmorTypeCommon    ArmorType = "common"    // Basic armors - Knight, Archer, Mage
	ArmorTypeRare      ArmorType = "rare"      // Rare armors - Light King, Light Emperor
	ArmorTypeLegendary ArmorType = "legendary" // Legendary armors - Only Emperor can buy
)

// Armor represents an armor in the system
type Armor struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Type        ArmorType          `bson:"type" json:"type"`
	Defense     int                `bson:"defense" json:"defense"`
	HPBonus     int                `bson:"hp_bonus" json:"hp_bonus"` // Additional HP provided by armor
	Price       int                `bson:"price" json:"price"`
	CreatedBy   string             `bson:"created_by" json:"created_by"` // warrior username
	OwnedBy     []string           `bson:"owned_by" json:"owned_by"`     // legacy: list of warrior usernames
	Durability  int                `bson:"durability" json:"durability"`
	MaxDurability int              `bson:"max_durability" json:"max_durability"`
	IsBroken    bool               `bson:"is_broken" json:"is_broken"`
	Owners      []OwnerRef         `bson:"owners,omitempty" json:"owners,omitempty"` // generalized ownership
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// OwnerRef for polymorphic ownership
type OwnerRef struct {
	OwnerType string `bson:"owner_type" json:"owner_type"` // warrior | enemy | dragon
	OwnerID   string `bson:"owner_id" json:"owner_id"`     // username or entity id
}

// CollectionName returns the MongoDB collection name
func (Armor) CollectionName() string {
	return "armors"
}

// CanBeCreatedBy checks if a role can create this armor type
func (at ArmorType) CanBeCreatedBy(role string) bool {
	// Only light emperor and light king can create armors
	if role != "light_emperor" && role != "light_king" {
		return false
	}

	// Legendary armors cannot be created by anyone
	if at == ArmorTypeLegendary {
		return false
	}

	return true
}

// CanBeBoughtBy checks if a role can buy this armor
func (a *Armor) CanBeBoughtBy(role string) bool {
	// Dark side cannot buy armors
	if role == "dark_emperor" || role == "dark_king" {
		return false
	}

	switch a.Type {
	case ArmorTypeCommon:
		// Common armors can be bought by knights, archers, mages, kings, emperors
		return true

	case ArmorTypeRare:
		// Rare armors can be bought by light king and light emperor
		return role == "light_king" || role == "light_emperor"

	case ArmorTypeLegendary:
		// Legendary armors can only be bought by emperors
		return role == "light_emperor"

	default:
		return false
	}
}

