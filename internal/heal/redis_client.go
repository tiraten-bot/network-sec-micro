package heal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedisClient initializes the Redis client connection
func InitRedisClient() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db = n
		}
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s (db=%d)", addr, db)
	return nil
}

// CloseRedisClient closes the Redis connection
func CloseRedisClient() {
	if redisClient != nil {
		redisClient.Close()
		log.Println("Redis client closed")
	}
}

// GetBattleLogLastHP retrieves the last HP value from battle logs for a warrior
func GetBattleLogLastHP(ctx context.Context, battleID string, warriorID uint) (int, error) {
	if redisClient == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	// Try battle:logs:{battleID} stream (latest entry)
	streamKey := fmt.Sprintf("battle:logs:%s", battleID)
	entries, err := redisClient.XRevRange(ctx, streamKey, "+", "-", 1).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	// Parse last entry to find warrior's final HP
	for _, entry := range entries {
		if data, ok := entry.Values["data"].(string); ok {
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(data), &logEntry); err == nil {
				// Check if this entry is for our warrior
				if attackerID, ok := logEntry["attacker_id"].(string); ok && attackerID == fmt.Sprintf("%d", warriorID) {
					if hpAfter, ok := logEntry["target_hp_after"].(float64); ok {
						return int(hpAfter), nil
					}
				}
				if targetID, ok := logEntry["target_id"].(string); ok && targetID == fmt.Sprintf("%d", warriorID) {
					if hpAfter, ok := logEntry["target_hp_after"].(float64); ok {
						return int(hpAfter), nil
					}
				}
			}
		}
	}

	// Fallback: try arena match from Redis
	arenaMatchKey := fmt.Sprintf("arena:match:%s", battleID)
	if data, err := redisClient.Get(ctx, arenaMatchKey).Bytes(); err == nil {
		var match map[string]interface{}
		if err := json.Unmarshal(data, &match); err == nil {
			// Check player1 or player2
			if p1ID, ok := match["player1_id"].(float64); ok && uint(p1ID) == warriorID {
				if hp, ok := match["player1_hp"].(float64); ok {
					return int(hp), nil
				}
			}
			if p2ID, ok := match["player2_id"].(float64); ok && uint(p2ID) == warriorID {
				if hp, ok := match["player2_hp"].(float64); ok {
					return int(hp), nil
				}
			}
		}
	}

	return 0, fmt.Errorf("could not find HP for warrior %d in battle %s", warriorID, battleID)
}

