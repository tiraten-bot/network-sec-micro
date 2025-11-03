package heal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealingLogEntry represents a healing process log entry in Redis
type HealingLogEntry struct {
	WarriorID    uint      `json:"warrior_id"`
	WarriorName  string    `json:"warrior_name"`
	HealType     string    `json:"heal_type"`
	Status       string    `json:"status"` // "started", "in_progress", "completed", "failed"
	HPBefore     int       `json:"hp_before"`
	HPAfter      int       `json:"hp_after"`
	HealedAmount int       `json:"healed_amount"`
	CoinsSpent   int       `json:"coins_spent"`
	Duration     int       `json:"duration"`      // Total duration in seconds
	Remaining    int       `json:"remaining"`     // Remaining seconds (for in_progress)
	Progress     float64   `json:"progress"`      // Progress percentage (0-100)
	Timestamp    time.Time `json:"timestamp"`
	Message      string    `json:"message,omitempty"`
}

// LogHealingStarted logs when healing starts
func LogHealingStarted(ctx context.Context, record *HealingRecord) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := HealingLogEntry{
		WarriorID:    record.WarriorID,
		WarriorName:  record.WarriorName,
		HealType:     string(record.HealType),
		Status:       "started",
		HPBefore:     record.HPBefore,
		HPAfter:      record.HPAfter,
		HealedAmount: record.HealedAmount,
		CoinsSpent:   record.CoinsSpent,
		Duration:     record.Duration,
		Remaining:    record.Duration,
		Progress:     0.0,
		Timestamp:    record.CreatedAt,
		Message:      fmt.Sprintf("Healing started: %s package purchased for %d coins, will heal %d HP in %d seconds", record.HealType, record.CoinsSpent, record.HealedAmount, record.Duration),
	}

	return logHealingEntry(ctx, record.WarriorID, logEntry)
}

// LogHealingProgress logs healing progress (periodic updates)
func LogHealingProgress(ctx context.Context, warriorID uint, warriorName string, healType HealType, remaining int, duration int, progress float64) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := HealingLogEntry{
		WarriorID:   warriorID,
		WarriorName: warriorName,
		HealType:    string(healType),
		Status:      "in_progress",
		Duration:    duration,
		Remaining:   remaining,
		Progress:    progress,
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Healing in progress: %.1f%% complete, %d seconds remaining", progress, remaining),
	}

	return logHealingEntry(ctx, warriorID, logEntry)
}

// LogHealingCompleted logs when healing completes
func LogHealingCompleted(ctx context.Context, record *HealingRecord) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := HealingLogEntry{
		WarriorID:    record.WarriorID,
		WarriorName:  record.WarriorName,
		HealType:     string(record.HealType),
		Status:       "completed",
		HPBefore:     record.HPBefore,
		HPAfter:      record.HPAfter,
		HealedAmount: record.HealedAmount,
		CoinsSpent:   record.CoinsSpent,
		Duration:     record.Duration,
		Remaining:    0,
		Progress:     100.0,
		Timestamp:    time.Now(),
		Message:      fmt.Sprintf("Healing completed: %d HP restored (from %d to %d)", record.HealedAmount, record.HPBefore, record.HPAfter),
	}

	return logHealingEntry(ctx, record.WarriorID, logEntry)
}

// LogHealingFailed logs when healing fails
func LogHealingFailed(ctx context.Context, warriorID uint, warriorName string, healType HealType, reason string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	logEntry := HealingLogEntry{
		WarriorID:   warriorID,
		WarriorName: warriorName,
		HealType:    string(healType),
		Status:      "failed",
		Timestamp:   time.Now(),
		Message:     fmt.Sprintf("Healing failed: %s", reason),
	}

	return logHealingEntry(ctx, warriorID, logEntry)
}

// logHealingEntry logs a healing entry to Redis Stream
func logHealingEntry(ctx context.Context, warriorID uint, entry HealingLogEntry) error {
	logData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Use Redis Stream for structured logging
	streamKey := fmt.Sprintf("healing:logs:%d", warriorID)

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

	// Also add to a sorted set for quick access by timestamp
	zsetKey := fmt.Sprintf("healing:logs:%d:timeline", warriorID)
	score := float64(entry.Timestamp.UnixNano())

	err = redisClient.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: string(logData),
	}).Err()

	if err != nil {
		// Don't fail the main operation
		return nil
	}

	// Set expiration for both keys (30 days - healing logs are important)
	expiration := 30 * 24 * time.Hour
	redisClient.Expire(ctx, streamKey, expiration)
	redisClient.Expire(ctx, zsetKey, expiration)

	return nil
}

// GetHealingLogs retrieves healing logs from Redis for a warrior
func GetHealingLogs(ctx context.Context, warriorID uint, limit int64) ([]HealingLogEntry, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	streamKey := fmt.Sprintf("healing:logs:%d", warriorID)

	// Get latest entries from stream
	entries, err := redisClient.XRevRangeN(ctx, streamKey, "+", "-", limit).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get logs from Redis: %w", err)
	}

	logs := make([]HealingLogEntry, 0, len(entries))
	for _, entry := range entries {
		if data, ok := entry.Values["data"].(string); ok {
			var logEntry HealingLogEntry
			if err := json.Unmarshal([]byte(data), &logEntry); err == nil {
				logs = append(logs, logEntry)
			}
		}
	}

	return logs, nil
}

