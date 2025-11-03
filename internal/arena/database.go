package arena

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var InvitationColl *mongo.Collection
var MatchColl *mongo.Collection

// InitDatabase initializes the MongoDB database connection
func InitDatabase() error {
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "arena_db")

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

	log.Println("MongoDB connection established for arena service")

	db := client.Database(dbName)
	InvitationColl = db.Collection((&ArenaInvitation{}).CollectionName())
	MatchColl = db.Collection((&ArenaMatch{}).CollectionName())

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Arena database initialized successfully")
	return nil
}

// createIndexes creates necessary indexes
func createIndexes() error {
	ctx := context.Background()

	invitationIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"challenger_id": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"opponent_id": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"expires_at": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1, "expires_at": 1},
		},
	}

	matchIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"player1_id": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"player2_id": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
	}

	_, err := InvitationColl.Indexes().CreateMany(ctx, invitationIndexes)
	if err != nil {
		return fmt.Errorf("failed to create invitation indexes: %w", err)
	}

	_, err = MatchColl.Indexes().CreateMany(ctx, matchIndexes)
	if err != nil {
		return fmt.Errorf("failed to create match indexes: %w", err)
	}

	log.Println("Arena indexes created successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

