package dto

// StartBattleRequest represents a request to start a battle
type StartBattleRequest struct {
	BattleType   string `json:"battle_type" binding:"required,oneof=enemy dragon"`
	OpponentID   string `json:"opponent_id" binding:"required"`
	MaxTurns     int    `json:"max_turns"` // Default 20 if not specified
}

// AttackRequest represents a request to perform an attack
type AttackRequest struct {
	BattleID string `json:"battle_id" binding:"required"`
}

