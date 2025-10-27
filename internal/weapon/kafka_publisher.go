package weapon

import (
	"context"
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

	log.Printf("Publishing weapon purchase event: Warrior %d purchased weapon %s for %d coins", 
		warriorID, weapon.Name, weapon.Price)

	// Get singleton Kafka publisher
	publisher, err := GetKafkaPublisher()
	if err != nil {
		log.Printf("Failed to get Kafka publisher: %v", err)
		return err
	}

	// Publish event to Kafka
	if err := publisher.Publish(kafka.TopicWeaponPurchase, event); err != nil {
		log.Printf("Failed to publish weapon purchase event to Kafka: %v", err)
		return err
	}

	log.Printf("Successfully published weapon purchase event to Kafka")
	return nil
}

