package heal

import (
	"encoding/json"
	"log"
	"os"

	"network-sec-micro/pkg/kafka"
)

// ProcessKafkaMessage processes incoming Kafka messages for healing events
func ProcessKafkaMessage(message []byte) error {
	// Try to unmarshal as arena match completed
	var arenaCompleted kafka.ArenaMatchCompletedEvent
	if err := json.Unmarshal(message, &arenaCompleted); err == nil {
		if arenaCompleted.Event.EventType == "arena_match_completed" {
			// Log that healing is available for both players
			if arenaCompleted.Player1ID > 0 {
				log.Printf("Arena match completed: Player1 (%d) can heal. Battle ID: %s", arenaCompleted.Player1ID, arenaCompleted.MatchID)
			}
			if arenaCompleted.Player2ID > 0 {
				log.Printf("Arena match completed: Player2 (%d) can heal. Battle ID: %s", arenaCompleted.Player2ID, arenaCompleted.MatchID)
			}
			return nil
		}
	}

	// Try to unmarshal as battle completed
	var battleCompleted struct {
		EventType    string `json:"event_type"`
		Timestamp    string `json:"timestamp"`
		SourceService string `json:"source_service"`
		BattleID     string `json:"battle_id"`
		BattleType   string `json:"battle_type"`
		Result       string `json:"result"`
		WinnerSide   string `json:"winner_side,omitempty"`
	}
	if err := json.Unmarshal(message, &battleCompleted); err == nil {
		if battleCompleted.EventType == "battle_completed" {
			log.Printf("Battle completed: Battle ID %s, result: %s. Participants can heal.", battleCompleted.BattleID, battleCompleted.Result)
			return nil
		}
	}

	log.Printf("Unknown event type or failed to process message in heal service")
	return nil
}

// StartKafkaConsumer starts consuming Kafka messages
func StartKafkaConsumer() error {
	brokers := getKafkaBrokers()
	consumer, err := kafka.NewConsumer(
		brokers,
		"heal-service-group",
		[]string{kafka.TopicArenaMatchCompleted, kafka.TopicBattleCompleted},
		ProcessKafkaMessage,
	)
	if err != nil {
		return err
	}

	log.Println("Heal service Kafka consumer started")
	if err := consumer.Start(); err != nil {
		return err
	}

	return nil
}

func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
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

