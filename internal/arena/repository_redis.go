package arena

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
)

type redisRepo struct{}

func arenaMatchKey(id string) string { return fmt.Sprintf("arena:match:%s", id) }

// GetMatchByID loads match snapshot from Redis.
func (r *redisRepo) GetMatchByID(ctx context.Context, id string) (*ArenaMatch, error) {
    rc := getRedis()
    if rc == nil { return nil, errors.New("redis not initialized") }
    data, err := rc.Get(ctx, arenaMatchKey(id)).Bytes()
    if err != nil { return nil, err }
    var m ArenaMatch
    if err := json.Unmarshal(data, &m); err != nil { return nil, err }
    return &m, nil
}

// UpdateMatchFields applies a partial update to the match stored as JSON in Redis.
func (r *redisRepo) UpdateMatchFields(ctx context.Context, id string, fields map[string]interface{}) error {
    rc := getRedis()
    if rc == nil { return errors.New("redis not initialized") }
    key := arenaMatchKey(id)
    // Load current
    data, err := rc.Get(ctx, key).Bytes()
    if err != nil { return err }
    var m map[string]interface{}
    if err := json.Unmarshal(data, &m); err != nil { return err }
    // Merge fields
    for k, v := range fields { m[k] = v }
    // Save
    enc, err := json.Marshal(m)
    if err != nil { return err }
    return rc.Set(ctx, key, enc, 0).Err()
}


