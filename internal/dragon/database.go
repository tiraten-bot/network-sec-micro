package dragon

import (
	"context"
	"fmt"
	"log"
	"time"

	"network-sec-micro/pkg/secrets"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client     *mongo.Client
	DB         *mongo.Database
	DragonColl *mongo.Collection
)

func InitDatabase() error {
	uri := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getEnv("MONGODB_DB", "dragon_db")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := Client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	DB = Client.Database(dbName)
	DragonColl = DB.Collection("dragons")

	log.Println("Dragon service database connection established")
	return nil
}

func getEnv(key, defaultValue string) string {
	return secrets.GetOrDefault(key, defaultValue)
}
