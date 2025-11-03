package dto

// PurchaseHealCommand represents a command to purchase a healing package
type PurchaseHealCommand struct {
	WarriorID   uint   `json:"warrior_id"`
	HealType    string `json:"heal_type"`    // "full", "partial", "emperor_full", "emperor_partial", "dragon"
	BattleID    string `json:"battle_id,omitempty"` // Optional battle ID for HP retrieval
	WarriorRole string `json:"warrior_role"` // Role for RBAC validation
}

