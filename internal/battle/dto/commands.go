package dto

// StartBattleCommand represents a command to start a new battle
type StartBattleCommand struct {
	BattleType    string `json:"battle_type" binding:"required,oneof=enemy dragon"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
	OpponentID    string `json:"opponent_id" binding:"required"`
	OpponentType  string `json:"opponent_type"`
	OpponentName  string `json:"opponent_name"`
	MaxTurns      int    `json:"max_turns"` // Maximum turns before draw
}

// AttackCommand represents a command to perform an attack in battle
type AttackCommand struct {
	BattleID      string `json:"battle_id" binding:"required"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
}

// CompleteBattleCommand represents a command to manually complete/cancel a battle
type CompleteBattleCommand struct {
	BattleID      string `json:"battle_id" binding:"required"`
	Reason        string `json:"reason"` // "victory", "defeat", "timeout", "cancelled"
}

