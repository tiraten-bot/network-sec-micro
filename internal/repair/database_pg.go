package repair

import (
	"fmt"

	"network-sec-micro/pkg/secrets"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitPostgres() error {
	host := secrets.GetOrDefault("DB_HOST", "localhost")
	port := secrets.GetOrDefault("DB_PORT", "5432")
	user := secrets.GetOrDefault("DB_USER", "postgres")
	pass := secrets.GetOrDefault("DB_PASSWORD", "postgres")
	name := secrets.GetOrDefault("DB_NAME_REPAIR", "repair_db")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return db.AutoMigrate(&RepairOrder{})
}

func GetDB() *gorm.DB { return db }
