package battle

import (
    "context"
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type mongoRepo struct{}

func (r *mongoRepo) GetBattleByID(ctx context.Context, id string) (*Battle, error) {
    if BattleColl == nil { return nil, errors.New("mongo not initialized") }
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { return nil, errors.New("invalid battle id") }
    var b Battle
    if err := BattleColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&b); err != nil {
        if err == mongo.ErrNoDocuments { return nil, errors.New("battle not found") }
        return nil, err
    }
    return &b, nil
}


