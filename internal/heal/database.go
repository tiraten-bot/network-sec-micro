package heal

import (
	"log"
	"os"

	pkgdb "network-sec-micro/pkg/database"
)

// SQLDB is the GORM PostgreSQL handle for Heal service
var SQLDB = struct {
	Enabled bool
	DB      interface{}
}{Enabled: false}

// InitPostgres initializes PostgreSQL via pkg/database
func InitPostgres() error {
	if os.Getenv("HEAL_USE_POSTGRES") == "" {
		return nil
	}
	cfg := pkgdb.GetConfigFromEnv()
	// allow override db name for heal
	if v := os.Getenv("DB_NAME_HEAL"); v != "" {
		cfg.DBName = v
	}
	db, err := pkgdb.Connect(cfg)
	if err != nil {
		return err
	}
	// AutoMigrate tables
	if err := pkgdb.AutoMigrate(db, &HealingRecordSQL{}); err != nil {
		return err
	}
	SQLDB.Enabled = true
	SQLDB.DB = db
	log.Println("Heal PostgreSQL initialized (GORM)")
	return nil
}

