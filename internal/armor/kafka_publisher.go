package armor

import (
	"context"
	"log"

	"network-sec-micro/pkg/kafka"
)

// PublishArmorPurchase publishes an armor purchase event to Kafka
func PublishArmorPurchase(ctx context.Context, armor *Armor, buyerID uint, buyerUsername string, ownerType string) error {
	// Create event
	event := kafka.NewArmorPurchaseEvent(
		armor.ID.Hex(),
		buyerUsername,
		armor.Name,
		int(buyerID),
		armor.Price,
		ownerType,
	)

	log.Printf("Publishing armor purchase event: %s %s purchased armor %s for %d coins", 
		ownerType, buyerUsername, armor.Name, armor.Price)

	// Get singleton Kafka publisher
	publisher, err := GetKafkaPublisher()
	if err != nil {
		log.Printf("Failed to get Kafka publisher: %v", err)
		return err
	}

	// Publish event to Kafka
	if err := publisher.Publish(kafka.TopicArmorPurchase, event); err != nil {
		log.Printf("Failed to publish armor purchase event to Kafka: %v", err)
		return err
	}

	log.Printf("Successfully published armor purchase event to Kafka")
	return nil
}

