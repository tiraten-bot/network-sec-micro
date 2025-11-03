package dto

// PurchaseHealCommand represents a command to purchase a healing package
type PurchaseHealCommand struct {
	ParticipantID   string `json:"participant_id"`   // Warrior/Dragon/Enemy ID (string for dragon/enemy)
	ParticipantType string `json:"participant_type"` // "warrior", "dragon", "enemy"
	HealType        string `json:"heal_type"`        // "full", "partial", "emperor_full", "emperor_partial", "dragon"
	BattleID        string `json:"battle_id,omitempty"` // Optional battle ID for HP retrieval
	ParticipantRole string `json:"participant_role"` // Role for RBAC validation
}

