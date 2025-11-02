package dto

// CastSpellRequest represents a request to cast a spell in battle
type CastSpellRequest struct {
	BattleID            string `json:"battle_id" binding:"required"`
	SpellType           string `json:"spell_type" binding:"required"` // call_of_the_light_king, resistance, rebirth, dragon_emperor, destroy_the_light, wraith_of_dragon
	TargetDragonID      string `json:"target_dragon_id,omitempty"`     // Required for Dragon Emperor spell
	TargetDarkEmperorID string `json:"target_dark_emperor_id,omitempty"` // Required for Dragon Emperor spell
}

