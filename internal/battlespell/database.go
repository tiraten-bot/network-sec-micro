package battlespell

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SpellColl *mongo.Collection

// InitDatabase initializes the MongoDB database connection
func InitDatabase() error {
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "battlespell_db")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("MongoDB connection established for battlespell service")

	db := client.Database(dbName)
	SpellColl = db.Collection((&Spell{}).CollectionName())

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Battlespell database initialized successfully")
	return nil
}

// createIndexes creates necessary indexes
func createIndexes() error {
	ctx := context.Background()

	spellIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"battle_id": 1, "spell_type": 1, "is_active": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1, "is_active": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1},
		},
	}

	_, err := SpellColl.Indexes().CreateMany(ctx, spellIndexes)
	if err != nil {
		return fmt.Errorf("failed to create spell indexes: %w", err)
	}

	log.Println("Battlespell indexes created successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

