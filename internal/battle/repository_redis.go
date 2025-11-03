package battle

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
)

type redisRepo struct{}

func battleKey(id string) string { return fmt.Sprintf("battle:%s", id) }

func (r *redisRepo) GetBattleByID(ctx context.Context, id string) (*Battle, error) {
    rc := GetRedisClient()
    if rc == nil { return nil, errors.New("redis not initialized") }
    data, err := rc.Get(ctx, battleKey(id)).Bytes()
    if err != nil { return nil, err }
    var b Battle
    if err := json.Unmarshal(data, &b); err != nil { return nil, err }
    return &b, nil
}


