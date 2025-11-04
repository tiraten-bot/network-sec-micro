package dto

// ParticipantInfo represents a participant to be added to battle
type ParticipantInfo struct {
	ParticipantID string `json:"participant_id" binding:"required"` // Warrior ID, Enemy ID, or Dragon ID
	Name          string `json:"name" binding:"required"`
	Type          string `json:"type" binding:"required"` // "warrior", "enemy", "dragon", "dark_king", "dark_emperor", "light_king", "light_emperor"
	Side          string `json:"side" binding:"required,oneof=light dark"` // light or dark
	Level         int    `json:"level"` // Level for hierarchy validation
	HP            int    `json:"hp"`
	MaxHP         int    `json:"max_hp"`
	AttackPower   int    `json:"attack_power"`
	Defense       int    `json:"defense"`
}

// StartBattleCommand represents a command to start a new team-based battle
type StartBattleCommand struct {
	LightSideName string           `json:"light_side_name"` // Optional: e.g., "Light Alliance"
	DarkSideName  string           `json:"dark_side_name"`  // Optional: e.g., "Dark Forces"
	LightParticipants []ParticipantInfo `json:"light_participants" binding:"required,min=1"` // At least 1 participant
	DarkParticipants  []ParticipantInfo `json:"dark_participants" binding:"required,min=1"`  // At least 1 participant
	MaxTurns      int              `json:"max_turns"` // Maximum turns before draw (default 100)
	CreatedBy     string           `json:"created_by"` // Creator username
    // Optional wager between emperors
    WagerAmount   int              `json:"wager_amount"`
    LightEmperorID string          `json:"light_emperor_id"`
    DarkEmperorID  string          `json:"dark_emperor_id"`
    RequireEmperorApproval bool    `json:"require_emperor_approval"`
    // Legacy single battle fields (optional)
    BattleType   string `json:"battle_type,omitempty"`
    WarriorName  string `json:"warrior_name,omitempty"`
    OpponentID   string `json:"opponent_id,omitempty"`
    OpponentName string `json:"opponent_name,omitempty"`
    OpponentType string `json:"opponent_type,omitempty"`
    OpponentHP   int    `json:"opponent_hp,omitempty"`
    OpponentMaxHP int   `json:"opponent_max_hp,omitempty"`
}

// AttackCommand represents a command to perform an attack in battle
type AttackCommand struct {
	BattleID      string `json:"battle_id" binding:"required"`
	AttackerID    string `json:"attacker_id" binding:"required"` // Participant ID making the attack
	TargetID      string `json:"target_id" binding:"required"`   // Participant ID being attacked
	AttackerName  string `json:"attacker_name"` // For validation
	TargetName    string `json:"target_name"`   // For validation
}

// AddParticipantCommand represents a command to add a participant to an existing battle (if still pending)
type AddParticipantCommand struct {
	BattleID      string           `json:"battle_id" binding:"required"`
	Participant   ParticipantInfo `json:"participant" binding:"required"`
}

// RemoveParticipantCommand represents a command to remove a participant from a pending battle
type RemoveParticipantCommand struct {
	BattleID      string `json:"battle_id" binding:"required"`
	ParticipantID string `json:"participant_id" binding:"required"`
}

// CompleteBattleCommand represents a command to manually complete/cancel a battle
type CompleteBattleCommand struct {
	BattleID      string `json:"battle_id" binding:"required"`
	Reason        string `json:"reason"` // "light_victory", "dark_victory", "draw", "cancelled"
}

// ReviveDragonCommand represents a command to revive a dragon in battle
type ReviveDragonCommand struct {
	BattleID           string `json:"battle_id" binding:"required"`
	DragonParticipantID string `json:"dragon_participant_id" binding:"required"`
}

// DarkEmperorJoinBattleCommand represents a command for Dark Emperor to join battle during crisis
type DarkEmperorJoinBattleCommand struct {
	BattleID            string `json:"battle_id" binding:"required"`
	DarkEmperorUsername string `json:"dark_emperor_username" binding:"required"`
	DarkEmperorUserID   string `json:"dark_emperor_user_id" binding:"required"`
	DragonParticipantID string `json:"dragon_participant_id" binding:"required"` // Required to check dragon status
}

// SacrificeDragonCommand represents a command to sacrifice dragon and revive enemies
type SacrificeDragonCommand struct {
	BattleID            string `json:"battle_id" binding:"required"`
	DragonParticipantID string `json:"dragon_participant_id" binding:"required"`
	DarkEmperorUsername string `json:"dark_emperor_username" binding:"required"`
}

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

