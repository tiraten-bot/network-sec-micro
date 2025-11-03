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
	ID           string    `json:"id"`
	WarriorID    uint      `json:"warrior_id"`
	WarriorName  string    `json:"warrior_name"`
	HealType     HealType  `json:"heal_type"`
	HealedAmount int       `json:"healed_amount"`
	HPBefore     int       `json:"hp_before"`
	HPAfter      int       `json:"hp_after"`
	CoinsSpent   int       `json:"coins_spent"`
	CreatedAt    time.Time `json:"created_at"`
}

// HealPackage represents pricing for healing packages
type HealPackage struct {
	Type        HealType `json:"type"`
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description"`
}

var (
	FullHealPackage = HealPackage{
		Type:        HealTypeFull,
		Name:        "Full Heal",
		Price:       100,
		Description: "Restore HP to maximum",
	}
	PartialHealPackage = HealPackage{
		Type:        HealTypePartial,
		Name:        "50% Heal",
		Price:       50,
		Description: "Restore 50% of current HP",
	}
)

