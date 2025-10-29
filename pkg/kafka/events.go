package kafka

import "time"

// Event represents a base event
type Event struct {
	EventType    string    `json:"event_type"`
	Timestamp    time.Time `json:"timestamp"`
	SourceService string   `json:"source_service"`
}

// WeaponPurchaseEvent represents a weapon purchase event
type WeaponPurchaseEvent struct {
	Event
	WeaponID     string `json:"weapon_id"`
	WarriorID    uint   `json:"warrior_id"`
	WarriorName  string `json:"warrior_name"`
	WeaponName   string `json:"weapon_name"`
	WeaponPrice  int    `json:"weapon_price"`
}

// NewWeaponPurchaseEvent creates a new weapon purchase event
func NewWeaponPurchaseEvent(weaponID, warriorName, weaponName string, warriorID, weaponPrice int) *WeaponPurchaseEvent {
	return &WeaponPurchaseEvent{
		Event: Event{
			EventType:     "weapon_purchased",
			Timestamp:     time.Now(),
			SourceService: "weapon",
		},
		WeaponID:    weaponID,
		WarriorID:   uint(warriorID),
		WarriorName: warriorName,
		WeaponName:  weaponName,
		WeaponPrice: weaponPrice,
	}
}

// DragonDeathEvent represents dragon death event for weapon loot
type DragonDeathEvent struct {
	EventType       string `json:"event_type"`
	Timestamp       string `json:"timestamp"`
	SourceService   string `json:"source_service"`
	DragonID        string `json:"dragon_id"`
	DragonName      string `json:"dragon_name"`
	DragonType      string `json:"dragon_type"`
	DragonLevel     int    `json:"dragon_level"`
	KillerUsername  string `json:"killer_username"`
	LootWeaponType  string `json:"loot_weapon_type"`
	LootWeaponName  string `json:"loot_weapon_name"`
}

// Topic names
const (
	TopicWeaponPurchase = "weapon.purchase"
	TopicCoinDeduct     = "coin.deduct"
	TopicDragonDeath    = "dragon.death"
)

