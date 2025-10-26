package warrior

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the PostgreSQL database connection
func InitDatabase() error {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "warrior_db")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Auto migrate the schema
	if err := DB.AutoMigrate(&Warrior{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed")

	// Seed initial data
	if err := seedDatabase(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

// seedDatabase creates initial warrior data
func seedDatabase() error {
	var count int64
	DB.Model(&Warrior{}).Count(&count)
	
	if count > 0 {
		return nil // Database already seeded
	}

	// Create test warriors
	warriors := []Warrior{
		{
			Username: "arthur",
			Email:    "arthur@kingdom.com",
			Password: hashPassword("king123"),
			Role:     RoleKing,
		},
		{
			Username: "lancelot",
			Email:    "lancelot@kingdom.com",
			Password: hashPassword("knight123"),
			Role:     RoleKnight,
		},
		{
			Username: "robin",
			Email:    "robin@kingdom.com",
			Password: hashPassword("archer123"),
			Role:     RoleArcher,
		},
		{
			Username: "merlin",
			Email:    "merlin@kingdom.com",
			Password: hashPassword("mage123"),
			Role:     RoleMage,
		},
	}

	if err := DB.Create(&warriors).Error; err != nil {
		return err
	}

	log.Println("Database seeded with initial warriors")
	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
