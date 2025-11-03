package heal

import "time"

// HealType represents the type of healing package
type HealType string

const (
	HealTypeFull           HealType = "full"            // Heal to max HP
	HealTypePartial        HealType = "partial"         // Heal 50% of current HP
	HealTypeEmperorFull    HealType = "emperor_full"    // Emperor full heal (fast, cheap)
	HealTypeEmperorPartial HealType = "emperor_partial" // Emperor partial heal (fast, cheap)
	HealTypeDragon         HealType = "dragon"          // Dragon heal (slow, expensive)
)

// HealingRecord represents a healing transaction
type HealingRecord struct {
	ID             string    `json:"id"`
	ParticipantID  string    `json:"participant_id"`  // Warrior/Dragon/Enemy ID (string for compatibility)
	ParticipantType string   `json:"participant_type"` // "warrior", "dragon", "enemy"
	ParticipantName string   `json:"participant_name"` // Name of the participant
	WarriorID      uint      `json:"warrior_id"`      // Legacy: kept for backward compatibility (0 for dragon/enemy)
	WarriorName    string    `json:"warrior_name"`    // Legacy: kept for backward compatibility
	HealType       HealType  `json:"heal_type"`
	HealedAmount   int       `json:"healed_amount"`
	HPBefore       int       `json:"hp_before"`
	HPAfter        int       `json:"hp_after"`
	CoinsSpent     int       `json:"coins_spent"`
	Duration       int       `json:"duration"`      // Healing duration in seconds
	CompletedAt    *time.Time `json:"completed_at"` // When healing completes
	CreatedAt      time.Time `json:"created_at"`
}

// HealPackage represents pricing and duration for healing packages
type HealPackage struct {
	Type        HealType `json:"type"`
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Duration    int      `json:"duration"` // Duration in seconds
	Description string   `json:"description"`
	RequiredRole string  `json:"required_role"` // Role required to use this package
}

var (
	FullHealPackage = HealPackage{
		Type:        HealTypeFull,
		Name:        "Full Heal",
		Price:       100,
		Duration:    300, // 5 minutes
		Description: "Restore HP to maximum",
		RequiredRole: "warrior",
	}
	PartialHealPackage = HealPackage{
		Type:        HealTypePartial,
		Name:        "50% Heal",
		Price:       50,
		Duration:    180, // 3 minutes
		Description: "Restore 50% of current HP",
		RequiredRole: "warrior",
	}
	EmperorFullHealPackage = HealPackage{
		Type:        HealTypeEmperorFull,
		Name:        "Emperor Full Heal",
		Price:       20,  // Cheap for emperors
		Duration:    30,  // 30 seconds - very fast
		Description: "Emperor exclusive: Fast full heal",
		RequiredRole: "emperor",
	}
	EmperorPartialHealPackage = HealPackage{
		Type:        HealTypeEmperorPartial,
		Name:        "Emperor Quick Heal",
		Price:       10,  // Very cheap
		Duration:    15,  // 15 seconds - very fast
		Description: "Emperor exclusive: Quick partial heal",
		RequiredRole: "emperor",
	}
	DragonHealPackage = HealPackage{
		Type:        HealTypeDragon,
		Name:        "Dragon Heal",
		Price:       1000, // Very expensive
		Duration:    3600, // 1 hour - very long
		Description: "Dragon exclusive: Powerful but slow heal",
		RequiredRole: "dragon",
	}
)

