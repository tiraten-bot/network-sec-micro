package coin

import (
	"fmt"
	"log"

	"network-sec-micro/pkg/secrets"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the MySQL database connection
func InitDatabase() error {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "coin_db")
	charset := getEnv("DB_CHARSET", "utf8mb4")
	parseTime := getEnv("DB_PARSE_TIME", "true")
	loc := getEnv("DB_LOC", "Local")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
		user, password, host, port, dbname, charset, parseTime, loc)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Coin service MySQL database connection established")

	// Auto migrate the schema
	if err := DB.AutoMigrate(&Transaction{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Coin service database migration completed")

	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	return secrets.GetOrDefault(key, defaultValue)
}
