package battle

import (
    "log"
    "os"

    pkgdb "network-sec-micro/pkg/database"
)

// SQLDB is the optional GORM PostgreSQL handle for Battle service
var SQLDB = struct{ Enabled bool; DB interface{} }{Enabled: false}

// InitPostgres initializes PostgreSQL via pkg/database when enabled by env
func InitPostgres() error {
    if os.Getenv("BATTLE_USE_POSTGRES") == "" {
        return nil
    }
    cfg := pkgdb.GetConfigFromEnv()
    // allow override db name for battle
    if v := os.Getenv("DB_NAME_BATTLE"); v != "" { cfg.DBName = v }
    db, err := pkgdb.Connect(cfg)
    if err != nil {
        return err
    }
    SQLDB.Enabled = true
    SQLDB.DB = db
    log.Println("Battle PostgreSQL initialized (GORM)")
    return nil
}


