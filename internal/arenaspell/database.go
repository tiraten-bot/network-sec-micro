package arenaspell

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
    Client    *mongo.Client
    DB        *mongo.Database
    SpellColl *mongo.Collection
)

// InitDatabase initializes MongoDB for arenaspell
func InitDatabase() error {
    uri := getEnv("MONGODB_URI", "mongodb://localhost:27017")
    dbName := getEnv("MONGODB_DB", "arenaspell_db")

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
    SpellColl = DB.Collection((Spell{}).CollectionName())

    log.Println("ArenaSpell MongoDB connection established")
    return nil
}

func getEnv(key, def string) string {
    return secrets.GetOrDefault(key, def)
}


