package arena

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"network-sec-micro/pkg/kafka"

	"github.com/IBM/sarama"
)

var kafkaConsumerGroup sarama.ConsumerGroup
var kafkaConsumerOnce sync.Once

// InitKafkaConsumer initializes the Kafka consumer group
func InitKafkaConsumer() error {
	var err error
	kafkaConsumerOnce.Do(func() {
		brokers := getKafkaBrokers()
		
		config := sarama.NewConfig()
		config.Version = sarama.V2_8_0_0
		config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		config.Consumer.Return.Errors = true

		consumerGroup, err := sarama.NewConsumerGroup(brokers, "arena-service", config)
		if err != nil {
			log.Printf("Failed to create Kafka consumer group: %v", err)
			return
		}

		kafkaConsumerGroup = consumerGroup
		log.Println("Kafka consumer group initialized for arena service")
	})
	return err
}

// CloseKafkaConsumer closes the Kafka consumer
func CloseKafkaConsumer() {
	if kafkaConsumerGroup != nil {
		kafkaConsumerGroup.Close()
	}
}

// ArenaConsumer represents the consumer for arena-related Kafka events
type ArenaConsumer struct {
	service *Service
}

// NewArenaConsumer creates a new arena consumer
func NewArenaConsumer(service *Service) *ArenaConsumer {
	return &ArenaConsumer{
		service: service,
	}
}

// StartConsuming starts consuming Kafka events
func (c *ArenaConsumer) StartConsuming(ctx context.Context) error {
	if kafkaConsumerGroup == nil {
		if err := InitKafkaConsumer(); err != nil {
			return fmt.Errorf("failed to initialize Kafka consumer: %w", err)
		}
	}

	handler := &arenaEventHandler{service: c.service}

	// Consume from battle completed topic to update arena matches
	go func() {
		for {
			if err := kafkaConsumerGroup.Consume(ctx, []string{kafka.TopicBattleCompleted}, handler); err != nil {
				log.Printf("Arena consumer error: %v", err)
				time.Sleep(5 * time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Handle errors
	go func() {
		for err := range kafkaConsumerGroup.Errors() {
			log.Printf("Arena Kafka consumer error: %v", err)
		}
	}()

	return nil
}

// arenaEventHandler handles Kafka events for arena service
type arenaEventHandler struct {
	service *Service
}

// Setup is called at the beginning of a new session, before ConsumeClaim
func (h *arenaEventHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called at the end of a session, once all ConsumeClaim goroutines have exited
func (h *arenaEventHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *arenaEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			topic := message.Topic
			switch topic {
			case kafka.TopicBattleCompleted:
				if err := h.handleBattleCompleted(message.Value); err != nil {
					log.Printf("Failed to handle battle completed event: %v", err)
				}
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleBattleCompleted handles battle completed events to update arena matches
func (h *arenaEventHandler) handleBattleCompleted(data []byte) error {
	// Parse battle completed event
	type BattleCompletedEvent struct {
		EventType      string `json:"event_type"`
		BattleID       string `json:"battle_id"`
		BattleType     string `json:"battle_type"`
		Result         string `json:"result"`
		WinnerName     string `json:"winner_name"`
		WinnerSide     string `json:"winner_side,omitempty"` // For team battles
		CoinsEarned    interface{} `json:"coins_earned"`
		ExperienceGained interface{} `json:"experience_gained"`
		TotalTurns     int    `json:"total_turns"`
	}

	var event BattleCompletedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal battle completed event: %w", err)
	}

	// Check if this is an arena battle (battle type might indicate it, or check if match exists)
	ctx := context.Background()
	
	// Find arena match by battle ID
	var match ArenaMatch
	err := MatchColl.FindOne(ctx, map[string]interface{}{
		"battle_id": event.BattleID,
	}).Decode(&match)

	if err != nil {
		// Not an arena match, ignore
		return nil
	}

	// Update match status
	now := time.Now()
	updateData := map[string]interface{}{
		"status":       MatchStatusCompleted,
		"completed_at": now,
		"updated_at":   now,
	}

	// Determine winner based on battle result
	// In arena battles, player1 is on light side, player2 is on dark side
	if event.Result == "light_victory" {
		// Player1 (light side) wins
		updateData["winner_id"] = match.Player1ID
		updateData["winner_name"] = match.Player1Name
	} else if event.Result == "dark_victory" {
		// Player2 (dark side) wins
		updateData["winner_id"] = match.Player2ID
		updateData["winner_name"] = match.Player2Name
	} else {
		// Draw - no winner
		updateData["winner_id"] = nil
		updateData["winner_name"] = ""
	}

	_, err = MatchColl.UpdateOne(ctx, map[string]interface{}{
		"_id": match.ID,
	}, map[string]interface{}{
		"$set": updateData,
	})

	if err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	// Get winner info for event
	var winnerID *uint
	var winnerName string
	if wid, ok := updateData["winner_id"].(uint); ok {
		winnerID = &wid
		winnerName = updateData["winner_name"].(string)
	}

	// Publish arena match completed event
	if err := PublishMatchCompleted(
		match.ID.Hex(),
		match.Player1ID,
		match.Player1Name,
		match.Player2ID,
		match.Player2Name,
		winnerID,
		winnerName,
		event.BattleID,
	); err != nil {
		log.Printf("Failed to publish match completed event: %v", err)
	}

	log.Printf("Arena match completed: %s vs %s, winner: %s", match.Player1Name, match.Player2Name, winnerName)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

