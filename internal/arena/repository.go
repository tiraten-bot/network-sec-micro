package arena

import (
    "context"
    "os"
)

// Repository abstracts persistence for Arena domain (Mongo or SQL)
type Repository interface {
    GetMatchByID(ctx context.Context, id string) (*ArenaMatch, error)
    UpdateMatchFields(ctx context.Context, id string, fields map[string]interface{}) error
    CreateMatch(ctx context.Context, m *ArenaMatch) (string, error)

    InsertInvitation(ctx context.Context, inv *ArenaInvitation) (string, error)
    GetInvitationByID(ctx context.Context, id string) (*ArenaInvitation, error)
    UpdateInvitationFields(ctx context.Context, id string, fields map[string]interface{}) error
    FindPendingInvitationBetween(ctx context.Context, challengerID, opponentID uint) (*ArenaInvitation, error)
}

var defaultRepo Repository

// GetRepository returns a singleton repo based on env (ARENA_STORE=sql|mongo)
func GetRepository() Repository {
    if defaultRepo != nil { return defaultRepo }
    store := os.Getenv("ARENA_STORE")
    switch store {
    case "redis":
        if getRedis() != nil {
            defaultRepo = &redisRepo{}
            return defaultRepo
        }
        // fallback to sql if redis not ready and sql enabled
        if SQLDB.Enabled { defaultRepo = &sqlRepo{} } else { defaultRepo = &redisRepo{} }
    case "sql":
        if SQLDB.Enabled {
            defaultRepo = &sqlRepo{}
            return defaultRepo
        }
        defaultRepo = &redisRepo{}
    default:
        if SQLDB.Enabled { defaultRepo = &sqlRepo{} } else if getRedis() != nil { defaultRepo = &redisRepo{} } else { defaultRepo = &sqlRepo{} }
    }
    return defaultRepo
}


