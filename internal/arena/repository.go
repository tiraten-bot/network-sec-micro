package arena

import (
    "context"
    "os"
)

// Repository abstracts persistence for Arena domain (Mongo or SQL)
type Repository interface {
    GetMatchByID(ctx context.Context, id string) (*ArenaMatch, error)
    UpdateMatchFields(ctx context.Context, id string, fields map[string]interface{}) error
}

var defaultRepo Repository

// GetRepository returns a singleton repo based on env (ARENA_STORE=sql|mongo)
func GetRepository() Repository {
    if defaultRepo != nil { return defaultRepo }
    store := os.Getenv("ARENA_STORE")
    if store == "sql" && SQLDB.Enabled {
        defaultRepo = &sqlRepo{}
    } else {
        defaultRepo = &mongoRepo{}
    }
    return defaultRepo
}


