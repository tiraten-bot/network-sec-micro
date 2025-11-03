package arena

import (
    "log"
    "os"

    pkgdb "network-sec-micro/pkg/database"
)

// SQLDB is the GORM PostgreSQL handle for Arena (optional, gradual migration)
var SQLDB = struct{ Enabled bool; DB interface{} }{Enabled: false}

// InitPostgres initializes PostgreSQL via pkg/database when enabled by env
func InitPostgres() error {
    if os.Getenv("ARENA_USE_POSTGRES") == "" {
        return nil
    }
    cfg := pkgdb.GetConfigFromEnv()
    // allow override db name for arena
    if v := os.Getenv("DB_NAME_ARENA"); v != "" { cfg.DBName = v }
    db, err := pkgdb.Connect(cfg)
    if err != nil {
        return err
    }
    // AutoMigrate relational models
    if err := pkgdb.AutoMigrate(db, &ArenaInvitationSQL{}, &ArenaMatchSQL{}); err != nil {
        return err
    }
    SQLDB.Enabled = true
    SQLDB.DB = db
    log.Println("Arena PostgreSQL initialized (GORM)")
    return nil
}


