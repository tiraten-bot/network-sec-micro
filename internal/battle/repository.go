package battle

import (
    "context"
    "os"
)

// Repository abstracts persistence for Battle domain
type Repository interface {
    GetBattleByID(ctx context.Context, id string) (*Battle, error)
}

var defaultRepo Repository

// GetRepository returns a singleton repo based on env (BATTLE_STORE=redis|mongo)
func GetRepository() Repository {
    if defaultRepo != nil { return defaultRepo }
    switch os.Getenv("BATTLE_STORE") {
    case "redis":
        if getRedis() != nil { defaultRepo = &redisRepo{}; return defaultRepo }
        defaultRepo = &mongoRepo{}
    default:
        defaultRepo = &mongoRepo{}
    }
    return defaultRepo
}


