package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SimpleBattleLogEntry represents a simplified battle log (only deaths and defeats)
type SimpleBattleLogEntry struct {
	BattleID      string    `json:"battle_id"`
	Timestamp     time.Time `json:"timestamp"`
	EventType     string    `json:"event_type"` // "participant_defeated", "team_defeated", "battle_started", "battle_ended"
	ParticipantID string    `json:"participant_id,omitempty"`
	ParticipantName string  `json:"participant_name,omitempty"`
	ParticipantType string  `json:"participant_type,omitempty"`
	ParticipantSide string  `json:"participant_side,omitempty"`
	KilledBy      []string  `json:"killed_by,omitempty"` // Array of participant IDs who contributed to kill
	Team          string    `json:"team,omitempty"` // "light" or "dark"
	Message       string    `json:"message"`
}

// LogParticipantDefeated logs when a participant is defeated in battle
func LogParticipantDefeated(ctx context.Context, battleID primitive.ObjectID, participant *BattleParticipant, killedByParticipants []*BattleParticipant) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	killedByIDs := make([]string, len(killedByParticipants))
	for i, p := range killedByParticipants {
		killedByIDs[i] = p.ParticipantID
	}

	logEntry := SimpleBattleLogEntry{
		BattleID:       battleID.Hex(),
		Timestamp:      time.Now(),
		EventType:      "participant_defeated",
		ParticipantID:  participant.ParticipantID,
		ParticipantName: participant.Name,
		ParticipantType: string(participant.Type),
		ParticipantSide: string(participant.Side),
		KilledBy:       killedByIDs,
		Message:        fmt.Sprintf("⚰️ %s (%s) öldürüldü", participant.Name, participant.Type),
	}

	// Add killer names to message
	if len(killedByParticipants) > 0 {
		killerNames := make([]string, len(killedByParticipants))
		for i, p := range killedByParticipants {
			killerNames[i] = p.Name
		}
		logEntry.Message = fmt.Sprintf("⚰️ %s (%s) şu birimler tarafından öldürüldü: %v", 
			participant.Name, participant.Type, killerNames)
	}

	logData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	streamKey := fmt.Sprintf("battle:logs:%s", battleID.Hex())
	_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": string(logData),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to write log to Redis: %w", err)
	}

	// Also add to sorted set
	zsetKey := fmt.Sprintf("battle:logs:%s:deaths", battleID.Hex())
	err = redisClient.ZAdd(ctx, zsetKey, redis.Z{
		Score:  float64(time.Now().UnixNano()),
		Member: string(logData),
	}).Err()

	if err != nil {
		log.Printf("Warning: failed to add to sorted set: %v", err)
	}

	// Set expiration (7 days)
	expiration := 7 * 24 * time.Hour
	redisClient.Expire(ctx, streamKey, expiration)
	redisClient.Expire(ctx, zsetKey, expiration)

	return nil
}

// LogBattleStart logs battle start (simplified)
func LogBattleStart(ctx context.Context, battle *Battle, message string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := SimpleBattleLogEntry{
		BattleID:   battle.ID.Hex(),
		Timestamp:  battle.CreatedAt,
		EventType:  "battle_started",
		Message:    message,
	}

	logData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	streamKey := fmt.Sprintf("battle:logs:%s", battle.ID.Hex())
	_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": string(logData),
		},
	}).Result()

	return err
}

// LogBattleEnd logs battle end (simplified)
func LogBattleEnd(ctx context.Context, battle *Battle, message string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := SimpleBattleLogEntry{
		BattleID:  battle.ID.Hex(),
		Timestamp: time.Now(),
		EventType: "battle_ended",
		Team:      string(battle.WinnerSide),
		Message:   message,
	}

	logData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	streamKey := fmt.Sprintf("battle:logs:%s", battle.ID.Hex())
	_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": string(logData),
		},
	}).Result()

	return err
}

// GetSimpleBattleLogs retrieves simplified battle logs (only deaths and important events)
func GetSimpleBattleLogs(ctx context.Context, battleID primitive.ObjectID, limit int64) ([]SimpleBattleLogEntry, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	streamKey := fmt.Sprintf("battle:logs:%s", battleID.Hex())
	
	messages, err := redisClient.XRevRangeN(ctx, streamKey, "+", "-", limit).Result()
	if err != nil {
		if err == redis.Nil {
			return []SimpleBattleLogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read logs from Redis: %w", err)
	}

	logs := make([]SimpleBattleLogEntry, 0, len(messages))
	for _, msg := range messages {
		var entry SimpleBattleLogEntry
		if data, ok := msg.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &entry); err != nil {
				log.Printf("Warning: failed to unmarshal log entry: %v", err)
				continue
			}
			logs = append(logs, entry)
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

