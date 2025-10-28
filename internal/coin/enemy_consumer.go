package coin

import (
	"context"
	"encoding/json"
	"log"

	pb "network-sec-micro/api/proto/coin"
)

// GoblinCoinStealEvent represents goblin coin steal event
type GoblinCoinStealEvent struct {
	EventType     string `json:"event_type"`
	Timestamp     string `json:"timestamp"`
	SourceService string `json:"source_service"`
	EnemyID       string `json:"enemy_id"`
	EnemyType     string `json:"enemy_type"`
	EnemyName     string `json:"enemy_name"`
	WarriorID     uint   `json:"warrior_id"`
	WarriorName   string `json:"warrior_name"`
	AttackType    string `json:"attack_type"`
	StolenValue   int    `json:"stolen_value"`
}

// ProcessEnemyAttackMessage processes enemy attack events from Kafka
func ProcessEnemyAttackMessage(message []byte) error {
	var event GoblinCoinStealEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal enemy attack event: %v", err)
		return err
	}

	// Only handle goblin coin steal events
	if event.AttackType != "coin_steal" || event.EnemyType != "goblin" {
		return nil
	}

	log.Printf("Processing goblin coin steal: %s stole %d coins from warrior %d", 
		event.EnemyName, event.StolenValue, event.WarriorID)

	// Deduct coins from warrior
	service := NewService()
	ctx := context.Background()

	if err := service.DeductCoins(dto.DeductCoinsCommand{
		WarriorID: event.WarriorID,
		Amount:    int64(event.StolenValue),
		Reason:    "goblin_attack: " + event.EnemyName + " stole your coins",
	}); err != nil {
		log.Printf("Failed to deduct coins from warrior %d: %v", event.WarriorID, err)
		return err
	}

	log.Printf("Successfully deducted %d coins from warrior %d", event.StolenValue, event.WarriorID)
	return nil
}

