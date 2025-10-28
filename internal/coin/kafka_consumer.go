package coin

import (
	"context"
	"encoding/json"
	"log"

	pb "network-sec-micro/api/proto/coin"
)

// WeaponPurchaseEvent represents the event structure
type WeaponPurchaseEvent struct {
	EventType     string `json:"event_type"`
	Timestamp     string `json:"timestamp"`
	SourceService string `json:"source_service"`
	WeaponID      string `json:"weapon_id"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
	WeaponName    string `json:"weapon_name"`
	WeaponPrice   int    `json:"weapon_price"`
}

// HandleWeaponPurchase handles weapon purchase events from Kafka
func (s *CoinServiceServer) HandleWeaponPurchase(event WeaponPurchaseEvent) error {
	log.Printf("Received weapon purchase event: %+v", event)

	// Deduct coins from warrior's balance
	ctx := context.Background()
	_, err := s.DeductCoins(ctx, &pb.DeductCoinsRequest{
		WarriorId: uint32(event.WarriorID),
		Amount:    int64(event.WeaponPrice),
		Reason:    "weapon_purchase: " + event.WeaponName,
	})

	if err != nil {
		log.Printf("Failed to deduct coins for warrior %d: %v", event.WarriorID, err)
		return err
	}

	log.Printf("Successfully deducted %d coins from warrior %d", event.WeaponPrice, event.WarriorID)
	return nil
}

// ProcessKafkaMessage processes incoming Kafka messages
func ProcessKafkaMessage(message []byte) error {
	// Try to unmarshal as weapon purchase event
	var weaponEvent WeaponPurchaseEvent
	if err := json.Unmarshal(message, &weaponEvent); err == nil {
		if weaponEvent.EventType == "weapon_purchased" {
			// Handle weapon purchase
			service := NewService()
			server := NewCoinServiceServer(service)
			return server.HandleWeaponPurchase(weaponEvent)
		}
	}

	// Try to unmarshal as enemy attack event
	if err := ProcessEnemyAttackMessage(message); err == nil {
		return nil // Successfully processed
	}

	log.Printf("Unknown event type or failed to process message")
	return nil
}

