package weapon

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB         *mongo.Database
	WeaponColl *mongo.Collection
)

// InitDatabase initializes the MongoDB connection
func InitDatabase() error {
	uri := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "weapon_db")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = client.Database(dbName)
	WeaponColl = DB.Collection("weapons")

	log.Println("MongoDB connection established")

	// Seed initial legendary weapons
	if err := seedDatabase(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

// seedDatabase creates initial legendary weapons
func seedDatabase() error {
	ctx := context.Background()

	// Check if weapons collection is empty
	count, err := WeaponColl.CountDocuments(ctx, nil)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Database already seeded
	}

	// Create legendary weapons (cannot be created by anyone, only bought)
	legendaryWeapons := []Weapon{
		{
			Name:        "Excalibur",
			Description: "The legendary sword of King Arthur",
			Type:        WeaponTypeLegendary,
			Damage:      1000,
			Price:       100000,
			CreatedBy:   "system",
			OwnedBy:     []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Dragon Slayer",
			Description: "A legendary weapon forged to slay dragons",
			Type:        WeaponTypeLegendary,
			Damage:      900,
			Price:       90000,
			CreatedBy:   "system",
			OwnedBy:     []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Lightbringer",
			Description: "The sword of the Light Emperor",
			Type:        WeaponTypeLegendary,
			Damage:      1200,
			Price:       150000,
			CreatedBy:   "system",
			OwnedBy:     []string{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, weapon := range legendaryWeapons {
		if _, err := WeaponColl.InsertOne(ctx, weapon); err != nil {
			return err
		}
	}

	log.Println("MongoDB seeded with legendary weapons")
	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
