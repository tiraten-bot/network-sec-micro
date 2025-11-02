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

// BattleLogEntry represents a single log entry for a battle turn
type BattleLogEntry struct {
	BattleID      string    `json:"battle_id"`
	TurnNumber    int       `json:"turn_number"`
	Timestamp     time.Time `json:"timestamp"`
	EventType     string    `json:"event_type"` // "warrior_attack", "opponent_attack", "critical_hit", "battle_start", "battle_end"
	AttackerID    string    `json:"attacker_id"`
	AttackerName  string    `json:"attacker_name"`
	AttackerType  string    `json:"attacker_type"` // "warrior" or "opponent"
	TargetID      string    `json:"target_id"`
	TargetName    string    `json:"target_name"`
	TargetType    string    `json:"target_type"`
	DamageDealt    int       `json:"damage_dealt"`
	CriticalHit    bool      `json:"critical_hit"`
	TargetHPBefore int       `json:"target_hp_before"`
	TargetHPAfter  int       `json:"target_hp_after"`
	WarriorHP      int       `json:"warrior_hp"`
	OpponentHP     int       `json:"opponent_hp"`
	Message        string    `json:"message,omitempty"`
}

// LogBattleTurn logs a battle turn to Redis Stream
func LogBattleTurn(ctx context.Context, battleID primitive.ObjectID, turn *BattleTurn, battle *Battle, eventType string, message string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := BattleLogEntry{
		BattleID:       battleID.Hex(),
		TurnNumber:     turn.TurnNumber,
		Timestamp:      turn.CreatedAt,
		EventType:      eventType,
		AttackerID:     turn.AttackerID,
		AttackerName:   turn.AttackerName,
		AttackerType:   turn.AttackerType,
		TargetID:       turn.TargetID,
		TargetName:     turn.TargetName,
		TargetType:     turn.TargetType,
		DamageDealt:    turn.DamageDealt,
		CriticalHit:    turn.CriticalHit,
		TargetHPAfter:  turn.TargetHPAfter,
		WarriorHP:      battle.WarriorHP,
		OpponentHP:     battle.OpponentHP,
		Message:        message,
	}

	// Determine TargetHPBefore based on target type
	// battle object passed should have HP before the attack
	if turn.TargetType == "warrior" {
		logEntry.TargetHPBefore = battle.WarriorHP
	} else {
		logEntry.TargetHPBefore = battle.OpponentHP
	}
	}

	// Marshal to JSON
	logData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Use Redis Stream for structured logging
	streamKey := fmt.Sprintf("battle:logs:%s", battleID.Hex())
	
	// Add entry to stream with auto-generated ID
	_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": string(logData),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to write log to Redis: %w", err)
	}

	// Also add to a sorted set for quick access by turn number
	zsetKey := fmt.Sprintf("battle:logs:%s:turns", battleID.Hex())
	score := float64(turn.TurnNumber*10000) + float64(turn.CreatedAt.UnixNano()%10000) // Unique score
	
	err = redisClient.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: string(logData),
	}).Err()
	
	if err != nil {
		log.Printf("Warning: failed to add to sorted set: %v", err)
		// Don't fail the main operation
	}

	// Set expiration for both keys (7 days)
	expiration := 7 * 24 * time.Hour
	redisClient.Expire(ctx, streamKey, expiration)
	redisClient.Expire(ctx, zsetKey, expiration)

	return nil
}

// LogBattleStart logs battle start event
func LogBattleStart(ctx context.Context, battle *Battle, message string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := BattleLogEntry{
		BattleID:     battle.ID.Hex(),
		TurnNumber:   0,
		Timestamp:    battle.CreatedAt,
		EventType:    "battle_start",
		AttackerName: battle.WarriorName,
		TargetName:   battle.OpponentName,
		WarriorHP:    battle.WarriorHP,
		OpponentHP:   battle.OpponentHP,
		Message:      message,
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

// LogBattleEnd logs battle end event
func LogBattleEnd(ctx context.Context, battle *Battle, message string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := BattleLogEntry{
		BattleID:     battle.ID.Hex(),
		TurnNumber:   battle.CurrentTurn,
		Timestamp:    time.Now(),
		EventType:    "battle_end",
		Message:      message,
		WarriorHP:    battle.WarriorHP,
		OpponentHP:   battle.OpponentHP,
	}

	if battle.WinnerName != "" {
		logEntry.Message = fmt.Sprintf("Battle ended. Winner: %s. %s", battle.WinnerName, message)
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

// GetBattleLogs retrieves battle logs from Redis
func GetBattleLogs(ctx context.Context, battleID primitive.ObjectID, limit int64) ([]BattleLogEntry, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	if limit <= 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Max limit
	}

	streamKey := fmt.Sprintf("battle:logs:%s", battleID.Hex())
	
	// Read from stream, starting from the beginning
	messages, err := redisClient.XRevRangeN(ctx, streamKey, "+", "-", limit).Result()
	if err != nil {
		if err == redis.Nil {
			return []BattleLogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read logs from Redis: %w", err)
	}

	logs := make([]BattleLogEntry, 0, len(messages))
	for _, msg := range messages {
		var entry BattleLogEntry
		if data, ok := msg.Values["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &entry); err != nil {
				log.Printf("Warning: failed to unmarshal log entry: %v", err)
				continue
			}
			logs = append(logs, entry)
		}
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

// GetBattleLogsByTurnRange retrieves battle logs for a specific turn range
func GetBattleLogsByTurnRange(ctx context.Context, battleID primitive.ObjectID, fromTurn, toTurn int) ([]BattleLogEntry, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	zsetKey := fmt.Sprintf("battle:logs:%s:turns", battleID.Hex())
	
	// Get logs from sorted set by score range
	fromScore := float64(fromTurn * 10000)
	toScore := float64((toTurn+1) * 10000)

	members, err := redisClient.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%.0f", fromScore),
		Max: fmt.Sprintf("%.0f", toScore),
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []BattleLogEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read logs from Redis: %w", err)
	}

	logs := make([]BattleLogEntry, 0, len(members))
	for _, member := range members {
		var entry BattleLogEntry
		if err := json.Unmarshal([]byte(member), &entry); err != nil {
			log.Printf("Warning: failed to unmarshal log entry: %v", err)
			continue
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

