package battle

import (
	"context"
	"fmt"
	"log"
	"time"

	"network-sec-micro/pkg/secrets"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var BattleColl *mongo.Collection
var BattleTurnColl *mongo.Collection
var BattleParticipantColl *mongo.Collection

// InitDatabase initializes the MongoDB database connection
func InitDatabase() error {
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "battle_db")

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

	log.Println("MongoDB connection established")

	db := client.Database(dbName)
	BattleColl = db.Collection((&Battle{}).CollectionName())
	BattleTurnColl = db.Collection((&BattleTurn{}).CollectionName())
	BattleParticipantColl = db.Collection((&BattleParticipant{}).CollectionName())

	// Create indexes
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Battle database initialized successfully")
	return nil
}

// createIndexes creates necessary indexes for battles
func createIndexes() error {
	ctx := context.Background()

	// Battle indexes
	battleIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"warrior_id": 1, "status": 1},
		},
		{
			Keys: map[string]interface{}{"opponent_id": 1, "battle_type": 1},
		},
		{
			Keys: map[string]interface{}{"created_at": -1},
		},
		{
			Keys: map[string]interface{}{"status": 1, "battle_type": 1},
		},
	}

	_, err := BattleColl.Indexes().CreateMany(ctx, battleIndexes)
	if err != nil {
		return fmt.Errorf("failed to create battle indexes: %w", err)
	}

	// BattleTurn indexes
	turnIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"battle_id": 1, "turn_number": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1},
		},
	}

	_, err = BattleTurnColl.Indexes().CreateMany(ctx, turnIndexes)
	if err != nil {
		return fmt.Errorf("failed to create battle_turn indexes: %w", err)
	}

	// BattleParticipant indexes
	participantIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"battle_id": 1, "side": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1, "is_alive": 1},
		},
		{
			Keys: map[string]interface{}{"battle_id": 1},
		},
		{
			Keys: map[string]interface{}{"participant_id": 1},
		},
	}

	_, err = BattleParticipantColl.Indexes().CreateMany(ctx, participantIndexes)
	if err != nil {
		return fmt.Errorf("failed to create battle_participant indexes: %w", err)
	}

	log.Println("Indexes created successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	return secrets.GetOrDefault(key, defaultValue)
}
