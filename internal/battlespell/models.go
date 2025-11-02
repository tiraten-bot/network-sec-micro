package battlespell

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamSide represents which side a participant is on
type TeamSide string

const (
	TeamSideLight TeamSide = "light"
	TeamSideDark  TeamSide = "dark"
)

// SpellType represents the type of spell
type SpellType string

const (
	// Light Spells
	SpellCallOfTheLightKing SpellType = "call_of_the_light_king" // Double attack power for all warriors
	SpellResistance         SpellType = "resistance"              // Double defense for all warriors
	SpellRebirth            SpellType = "rebirth"                 // Revive all warriors

	// Dark Spells
	SpellDragonEmperor   SpellType = "dragon_emperor"   // Add Dark Emperor stats to dragon
	SpellDestroyTheLight SpellType = "destroy_the_light" // Reduce warrior attack/defense by 30% (stackable up to 2 times)
	SpellWraithOfDragon  SpellType = "wraith_of_dragon"  // When dragon kills warrior, random warrior also dies (max 25 times)
)

// SpellSide indicates which side the spell belongs to
func (st SpellType) Side() TeamSide {
	switch st {
	case SpellCallOfTheLightKing, SpellResistance, SpellRebirth:
		return TeamSideLight
	case SpellDragonEmperor, SpellDestroyTheLight, SpellWraithOfDragon:
		return TeamSideDark
	default:
		return ""
	}
}

// Spell represents a spell cast during battle
type Spell struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BattleID      primitive.ObjectID `bson:"battle_id" json:"battle_id"`
	SpellType     SpellType          `bson:"spell_type" json:"spell_type"`
	Side          TeamSide           `bson:"side" json:"side"`

	// Caster info
	CasterUsername string `bson:"caster_username" json:"caster_username"`
	CasterUserID   string `bson:"caster_user_id" json:"caster_user_id"`
	CasterRole     string `bson:"caster_role" json:"caster_role"` // light_king or dark_king

	// Target info (if applicable)
	TargetDragonID       string `bson:"target_dragon_id,omitempty" json:"target_dragon_id,omitempty"`             // For Dragon Emperor spell
	TargetDarkEmperorID  string `bson:"target_dark_emperor_id,omitempty" json:"target_dark_emperor_id,omitempty"` // For Dragon Emperor spell

	// Effect tracking
	StackCount  int `bson:"stack_count,omitempty" json:"stack_count,omitempty"`   // For Destroy the Light (max 2)
	WraithCount int `bson:"wraith_count,omitempty" json:"wraith_count,omitempty"` // For Wraith of Dragon (max 25)

	// Status
	IsActive bool      `bson:"is_active" json:"is_active"`
	CastAt   time.Time `bson:"cast_at" json:"cast_at"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (Spell) CollectionName() string {
	return "battle_spells"
}

// CanBeCastBy checks if a role can cast this spell
func (st SpellType) CanBeCastBy(role string) bool {
	side := st.Side()
	if side == TeamSideLight {
		return role == "light_king"
	} else if side == TeamSideDark {
		return role == "dark_king"
	}
	return false
}

// IsLightSpell checks if spell is for light side
func (st SpellType) IsLightSpell() bool {
	return st.Side() == TeamSideLight
}

// IsDarkSpell checks if spell is for dark side
func (st SpellType) IsDarkSpell() bool {
	return st.Side() == TeamSideDark
}

