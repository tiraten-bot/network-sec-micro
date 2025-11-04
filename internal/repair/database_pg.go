package repair

import (
    "fmt"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var db *gorm.DB

func InitPostgres() error {
    host := os.Getenv("DB_HOST"); port := os.Getenv("DB_PORT"); user := os.Getenv("DB_USER"); pass := os.Getenv("DB_PASSWORD"); name := os.Getenv("DB_NAME_REPAIR")
    if host == "" { host = "localhost" }
    if port == "" { port = "5432" }
    if user == "" { user = "postgres" }
    if pass == "" { pass = "postgres" }
    if name == "" { name = "repair_db" }
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil { return err }
    return db.AutoMigrate(&RepairOrder{})
}

func GetDB() *gorm.DB { return db }


