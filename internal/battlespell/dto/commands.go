package dto

// CastSpellCommand represents a command to cast a spell in battle
type CastSpellCommand struct {
	BattleID            string `json:"battle_id"`
	SpellType           string `json:"spell_type"`
	CasterUsername      string `json:"caster_username"`
	CasterUserID        string `json:"caster_user_id"`
	CasterRole          string `json:"caster_role"` // light_king or dark_king
	TargetDragonID      string `json:"target_dragon_id,omitempty"`     // Required for Dragon Emperor spell
	TargetDarkEmperorID string `json:"target_dark_emperor_id,omitempty"` // Required for Dragon Emperor spell
}

