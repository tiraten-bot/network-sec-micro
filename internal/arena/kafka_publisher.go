package arena

import (
	"fmt"
	"log"
	"os"
	"sync"

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
		log.Println("Kafka publisher initialized for arena service")
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

// PublishInvitationSent publishes invitation sent event
func PublishInvitationSent(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string, expiresAt string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaInvitationSentEvent(invitationID, challengerID, challengerName, opponentID, opponentName, expiresAt)
	topic := kafka.TopicArenaInvitationSent
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish invitation sent event: %w", err)
	}

	log.Printf("Published arena invitation sent event: %s -> %s", challengerName, opponentName)
	return nil
}

// PublishInvitationAccepted publishes invitation accepted event
func PublishInvitationAccepted(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string, battleID string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaInvitationAcceptedEvent(invitationID, challengerID, challengerName, opponentID, opponentName, battleID)
	topic := kafka.TopicArenaInvitationAccepted
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish invitation accepted event: %w", err)
	}

	log.Printf("Published arena invitation accepted event: %s vs %s", challengerName, opponentName)
	return nil
}

// PublishInvitationRejected publishes invitation rejected event
func PublishInvitationRejected(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaInvitationRejectedEvent(invitationID, challengerID, challengerName, opponentID, opponentName)
	topic := kafka.TopicArenaInvitationRejected
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish invitation rejected event: %w", err)
	}

	log.Printf("Published arena invitation rejected event: %s -> %s (rejected)", challengerName, opponentName)
	return nil
}

// PublishInvitationExpired publishes invitation expired event
func PublishInvitationExpired(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaInvitationExpiredEvent(invitationID, challengerID, challengerName, opponentID, opponentName)
	topic := kafka.TopicArenaInvitationExpired
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish invitation expired event: %w", err)
	}

	log.Printf("Published arena invitation expired event: %s -> %s", challengerName, opponentName)
	return nil
}

// PublishMatchStarted publishes arena match started event
func PublishMatchStarted(matchID string, player1ID uint, player1Name string, player2ID uint, player2Name string, battleID string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaMatchStartedEvent(matchID, player1ID, player1Name, player2ID, player2Name, battleID)
	topic := kafka.TopicArenaMatchStarted
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish match started event: %w", err)
	}

	log.Printf("Published arena match started event: %s vs %s", player1Name, player2Name)
	return nil
}

// PublishMatchCompleted publishes arena match completed event
func PublishMatchCompleted(matchID string, player1ID uint, player1Name string, player2ID uint, player2Name string, winnerID *uint, winnerName string, battleID string) error {
	publisher := GetKafkaPublisher()
	if publisher == nil {
		return fmt.Errorf("kafka publisher not initialized")
	}

	event := kafka.NewArenaMatchCompletedEvent(matchID, player1ID, player1Name, player2ID, player2Name, winnerID, winnerName, battleID)
	topic := kafka.TopicArenaMatchCompleted
	if err := publisher.Publish(topic, event); err != nil {
		return fmt.Errorf("failed to publish match completed event: %w", err)
	}

	log.Printf("Published arena match completed event: %s vs %s, winner: %s", player1Name, player2Name, winnerName)
	return nil
}

