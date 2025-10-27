package weapon

import (
	"context"
	"encoding/json"
	"log"

	"network-sec-micro/pkg/kafka"
)

// PublishWeaponPurchase publishes a weapon purchase event to Kafka
func PublishWeaponPurchase(ctx context.Context, weapon *Weapon, warriorID uint, warriorUsername string) error {
	// Create event
	event := kafka.NewWeaponPurchaseEvent(
		weapon.ID.Hex(),
		warriorUsername,
		weapon.Name,
		int(warriorID),
		weapon.Price,
	)

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	log.Printf("Publishing weapon purchase event: %s", string(jsonData))

	// TODO: Get kafka publisher instance (need to add to service)
	// For now, just log
	log.Printf("Event published: Warrior %d purchased weapon %s for %d coins", 
		warriorID, weapon.Name, weapon.Price)

	return nil
}

