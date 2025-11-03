package arenaspell

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// SpellType represents the type of arena spell (1v1 oriented)
type SpellType string

const (
    // Light-side inspired buffs (self only in 1v1)
    SpellCallOfTheLightKing SpellType = "call_of_the_light_king" // Double attack of caster
    SpellResistance         SpellType = "resistance"              // Double defense of caster
    SpellRebirth            SpellType = "rebirth"                 // Revive caster to 50% HP once

    // Dark-side inspired debuff (applies to opponent in 1v1)
    SpellDestroyTheLight SpellType = "destroy_the_light" // Reduce opponent attack/defense by 30% (stack up to 2)

    // Crisis spells (both sides) - unlocked at critical HP window (< =10%)
    SpellLightCrisis SpellType = "light_crisis" // Strong light-side emergency buff
    SpellDarkCrisis  SpellType = "dark_crisis"  // Strong dark-side emergency debuff
)

type TeamSide string

const (
    TeamSideLight TeamSide = "light"
    TeamSideDark  TeamSide = "dark"
)

func (st SpellType) Side() TeamSide {
    switch st {
    case SpellCallOfTheLightKing, SpellResistance, SpellRebirth, SpellLightCrisis:
        return TeamSideLight
    case SpellDestroyTheLight, SpellDarkCrisis:
        return TeamSideDark
    default:
        return ""
    }
}

// CanBeCastBy validates role-based casting (kings only by side)
func (st SpellType) CanBeCastBy(role string) bool {
    side := st.Side()
    if side == TeamSideLight {
        return role == "light_king"
    } else if side == TeamSideDark {
        return role == "dark_king"
    }
    return false
}

// Spell represents a spell cast in an arena match
type Spell struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MatchID   primitive.ObjectID `bson:"match_id" json:"match_id"`
    SpellType SpellType          `bson:"spell_type" json:"spell_type"`

    // Caster info
    CasterUserID   uint   `bson:"caster_user_id" json:"caster_user_id"`
    CasterUsername string `bson:"caster_username" json:"caster_username"`

    // Stack tracking for debuffs
    StackCount int `bson:"stack_count,omitempty" json:"stack_count,omitempty"`

    // Status
    IsActive  bool      `bson:"is_active" json:"is_active"`
    CastAt    time.Time `bson:"cast_at" json:"cast_at"`
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name for arena spells
func (Spell) CollectionName() string {
    return "arena_spells"
}


