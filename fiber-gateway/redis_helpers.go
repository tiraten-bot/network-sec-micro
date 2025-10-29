package main

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

func redisAllow(ctx context.Context, rdb *redis.Client, key string, limit int, windowSec int) (bool, error) {
    n, err := rdb.Incr(ctx, key).Result()
    if err != nil { return false, err }
    if n == 1 { _ = rdb.Expire(ctx, key, time.Duration(windowSec)*time.Second).Err() }
    return n <= int64(limit), nil
}

func redisQuota(ctx context.Context, rdb *redis.Client, key string, limit int, ttlSec int) (bool, error) {
    n, err := rdb.Incr(ctx, key).Result()
    if err != nil { return false, err }
    if n == 1 { _ = rdb.Expire(ctx, key, time.Duration(ttlSec)*time.Second).Err() }
    return n <= int64(limit), nil
}


