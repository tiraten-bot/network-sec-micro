package dragon

import (
	"encoding/json"
	"fmt"
	"log"

	"network-sec-micro/pkg/kafka"
)

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

// PublishDragonDeathEvent publishes dragon death event
func PublishDragonDeathEvent(event DragonDeathEvent) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal dragon death event: %w", err)
	}

	topic := kafka.TopicDragonDeath
	if err := publisher.PublishMessage(topic, message); err != nil {
		return fmt.Errorf("failed to publish dragon death event: %w", err)
	}

	log.Printf("Published dragon death event: %s killed %s (level %d)", 
		event.KillerUsername, event.DragonName, event.DragonLevel)
	return nil
}
