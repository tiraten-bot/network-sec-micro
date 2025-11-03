package arena

import (
    "context"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedisClient initializes the Redis client connection for Arena service
func InitRedisClient() error {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "localhost:6379"
    }

    password := os.Getenv("REDIS_PASSWORD")
    db := 0
    if v := os.Getenv("REDIS_DB"); v != "" {
        if n, err := strconv.Atoi(v); err == nil { db = n }
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

    log.Printf("[arena] Connected to Redis at %s (db=%d)", addr, db)
    return nil
}

// CloseRedisClient closes the Redis connection
func CloseRedisClient() {
    if redisClient != nil {
        _ = redisClient.Close()
        log.Println("[arena] Redis client closed")
    }
}

// getRedis returns the Redis client instance
func getRedis() *redis.Client { return redisClient }


