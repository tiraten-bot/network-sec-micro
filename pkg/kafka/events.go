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

// Topic names
const (
	TopicWeaponPurchase = "weapon.purchase"
	TopicCoinDeduct     = "coin.deduct"
)

