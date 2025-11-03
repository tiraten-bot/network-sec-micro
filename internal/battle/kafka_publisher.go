package battle

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"network-sec-micro/pkg/kafka"
)

var kafkaPublisher *kafka.Publisher
var kafkaPublisherOnce sync.Once

// GetKafkaPublisher returns the Kafka publisher singleton
func GetKafkaPublisher() *kafka.Publisher {
	kafkaPublisherOnce.Do(func() {
		brokers := getKafkaBrokers()
		publisher, err := kafka.NewPublisher(brokers)
		if err != nil {
			log.Printf("Warning: Failed to initialize Kafka publisher: %v", err)
			return
		}
		kafkaPublisher = publisher
		log.Println("Kafka publisher initialized")
	})
	return kafkaPublisher
}

// CloseKafkaPublisher closes the Kafka publisher
func CloseKafkaPublisher() {
	if kafkaPublisher != nil {
		kafkaPublisher.Close()
	}
}

func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	// Simple split by comma
	var result []string
	for _, b := range splitAndTrim(brokers, ",") {
		if b != "" {
			result = append(result, b)
		}
	}
	if len(result) == 0 {
		return []string{"localhost:9092"}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || string(s[i]) == sep {
			part := s[start:i]
			// trim spaces
			for len(part) > 0 && (part[0] == ' ' || part[0] == '\t') {
				part = part[1:]
			}
			for len(part) > 0 && (part[len(part)-1] == ' ' || part[len(part)-1] == '\t') {
				part = part[:len(part)-1]
			}
			parts = append(parts, part)
			start = i + 1
		}
	}
	return parts
}

// BattleStartedEvent represents a battle started event
type BattleStartedEvent struct {
	EventType      string    `json:"event_type"`
	Timestamp      time.Time `json:"timestamp"`
	SourceService  string    `json:"source_service"`
	BattleID       string    `json:"battle_id"`
	BattleType     string    `json:"battle_type"`
	WarriorID      uint      `json:"warrior_id"`
	WarriorName    string    `json:"warrior_name"`
	OpponentID     string    `json:"opponent_id"`
	OpponentName   string    `json:"opponent_name"`
	OpponentType    string    `json:"opponent_type"`
}

// BattleCompletedEvent represents a battle completed event
type BattleCompletedEvent struct {
	EventType         string    `json:"event_type"`
	Timestamp         time.Time `json:"timestamp"`
	SourceService     string    `json:"source_service"`
	BattleID          string    `json:"battle_id"`
	BattleType        string    `json:"battle_type"`
	WarriorID         uint      `json:"warrior_id"`
	WarriorName       string    `json:"warrior_name"`
	Result            string    `json:"result"`
	WinnerName        string    `json:"winner_name,omitempty"`
	CoinsEarned       int       `json:"coins_earned,omitempty"`
	ExperienceGained  int       `json:"experience_gained,omitempty"`
	TotalTurns        int       `json:"total_turns"`
}

// PublishBattleStartedEvent publishes battle started event
func PublishBattleStartedEvent(battleID string, battleType BattleType, warriorID uint, warriorName, opponentID, opponentName, opponentType string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := BattleStartedEvent{
		EventType:     "battle_started",
		Timestamp:     time.Now(),
		SourceService: "battle",
		BattleID:      battleID,
		BattleType:    string(battleType),
		WarriorID:     warriorID,
		WarriorName:   warriorName,
		OpponentID:    opponentID,
		OpponentName:  opponentName,
		OpponentType:  opponentType,
	}

	topic := kafka.TopicBattleStarted
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish battle started event: %w", err)
	}

	log.Printf("Published battle started event: %s vs %s", warriorName, opponentName)
	return nil
}

// PublishBattleCompletedEvent publishes battle completed event
func PublishBattleCompletedEvent(battleID string, battleType BattleType, warriorID uint, warriorName, result, winnerName string, coinsEarned, experienceGained, totalTurns int) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := BattleCompletedEvent{
		EventType:        "battle_completed",
		Timestamp:        time.Now(),
		SourceService:    "battle",
		BattleID:         battleID,
		BattleType:       string(battleType),
		WarriorID:        warriorID,
		WarriorName:      warriorName,
		Result:           result,
		WinnerName:       winnerName,
		CoinsEarned:      coinsEarned,
		ExperienceGained: experienceGained,
		TotalTurns:       totalTurns,
	}

	topic := kafka.TopicBattleCompleted
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish battle completed event: %w", err)
	}

	log.Printf("Published battle completed event: %s - %s", battleID, result)
	return nil
}

// PublishBattleWagerResolved publishes wager resolution after team battle
func PublishBattleWagerResolved(battleID string, winnerSide TeamSide, wagerAmount int, lightEmperorID, darkEmperorID string) error {
    publisher := GetKafkaPublisher()
    if publisher == nil { return fmt.Errorf("kafka publisher not initialized") }
    event := kafka.NewBattleWagerResolvedEvent(battleID, string(winnerSide), wagerAmount, lightEmperorID, darkEmperorID)
    topic := kafka.TopicBattleWagerResolved
    if err := publisher.Publish(topic, event); err != nil { return fmt.Errorf("failed to publish battle wager resolved: %w", err) }
    log.Printf("Published battle wager resolved: battle=%s winner=%s amount=%d", battleID, winnerSide, wagerAmount)
    return nil
}

