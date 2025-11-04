package armor

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
	DB        *mongo.Database
	ArmorColl *mongo.Collection
)

// InitDatabase initializes the MongoDB connection
func InitDatabase() error {
	uri := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "armor_db")

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
	ArmorColl = DB.Collection("armors")

	log.Println("MongoDB connection established for armor service")

	// Seed initial legendary armors
	if err := seedDatabase(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	return nil
}

// seedDatabase creates initial legendary armors
func seedDatabase() error {
	ctx := context.Background()

	// Check if armors collection is empty
	count, err := ArmorColl.CountDocuments(ctx, nil)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Database already seeded
	}

	// Create legendary armors (cannot be created by anyone, only bought)
	legendaryArmors := []Armor{
		{
			Name:         "Dragon Scale Armor",
			Description:  "The legendary armor forged from dragon scales",
			Type:         ArmorTypeLegendary,
			Defense:      500,
			HPBonus:      1000,
			Price:        100000,
			CreatedBy:    "system",
			OwnedBy:      []string{},
			Durability:   1000,
			MaxDurability: 1000,
			IsBroken:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Lightbringer Plate",
			Description:  "The legendary plate armor of the Light Emperor",
			Type:         ArmorTypeLegendary,
			Defense:      600,
			HPBonus:      1200,
			Price:        150000,
			CreatedBy:    "system",
			OwnedBy:      []string{},
			Durability:   1200,
			MaxDurability: 1200,
			IsBroken:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Immortal Guard",
			Description:  "A legendary armor that grants near-immortality",
			Type:         ArmorTypeLegendary,
			Defense:      700,
			HPBonus:      1500,
			Price:        200000,
			CreatedBy:    "system",
			OwnedBy:      []string{},
			Durability:   1500,
			MaxDurability: 1500,
			IsBroken:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, armor := range legendaryArmors {
		if _, err := ArmorColl.InsertOne(ctx, armor); err != nil {
			return err
		}
	}

	log.Println("MongoDB seeded with legendary armors")
	return nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

