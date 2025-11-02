package dto

// StartBattleRequest represents a request to start a team battle
type StartBattleRequest struct {
	LightSideName      string           `json:"light_side_name" binding:"required"` // e.g., "Light Alliance"
	DarkSideName       string           `json:"dark_side_name" binding:"required"`   // e.g., "Dark Forces"
	LightParticipants  []ParticipantInfo `json:"light_participants" binding:"required,min=1,dive"`
	DarkParticipants   []ParticipantInfo `json:"dark_participants" binding:"required,min=1,dive"`
	MaxTurns           int              `json:"max_turns"` // Default 100 if not specified
	KingApprovals      []uint           `json:"king_approvals,omitempty"` // List of king IDs who approved (required if creator is a king)
}

// AttackRequest represents a request to perform an attack
type AttackRequest struct {
	BattleID   string `json:"battle_id" binding:"required"`
	AttackerID string `json:"attacker_id" binding:"required"` // Participant ID
	TargetID   string `json:"target_id" binding:"required"`   // Participant ID
}

// AddParticipantRequest represents a request to add a participant to battle (pending only)
type AddParticipantRequest struct {
	BattleID    string          `json:"battle_id" binding:"required"`
	Participant ParticipantInfo `json:"participant" binding:"required"`
}

// RemoveParticipantRequest represents a request to remove a participant from battle (pending only)
type RemoveParticipantRequest struct {
	BattleID     string `json:"battle_id" binding:"required"`
	ParticipantID string `json:"participant_id" binding:"required"`
}

// ReviveDragonRequest represents a request to revive a dragon in battle
type ReviveDragonRequest struct {
	BattleID           string `json:"battle_id" binding:"required"`
	DragonParticipantID string `json:"dragon_participant_id" binding:"required"`
}

// DarkEmperorJoinBattleRequest represents a request for Dark Emperor to join battle during crisis
type DarkEmperorJoinBattleRequest struct {
	BattleID            string `json:"battle_id" binding:"required"`
	DarkEmperorUsername string `json:"dark_emperor_username" binding:"required"`
	DarkEmperorUserID   string `json:"dark_emperor_user_id" binding:"required"`
}

// SacrificeDragonRequest represents a request to sacrifice dragon and revive enemies
type SacrificeDragonRequest struct {
	BattleID            string `json:"battle_id" binding:"required"`
	DragonParticipantID string `json:"dragon_participant_id" binding:"required"`
	DarkEmperorUsername string `json:"dark_emperor_username" binding:"required"`
}
