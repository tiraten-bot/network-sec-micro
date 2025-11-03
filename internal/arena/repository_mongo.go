package arena

import (
    "context"
    "errors"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoRepo struct{}

func (r *mongoRepo) GetMatchByID(ctx context.Context, id string) (*ArenaMatch, error) {
    if MatchColl == nil { return nil, errors.New("mongo not initialized") }
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { return nil, errors.New("invalid match id") }
    var m ArenaMatch
    if err := MatchColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&m); err != nil { return nil, err }
    return &m, nil
}

func (r *mongoRepo) UpdateMatchFields(ctx context.Context, id string, fields map[string]interface{}) error {
    if MatchColl == nil { return errors.New("mongo not initialized") }
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { return errors.New("invalid match id") }
    _, err = MatchColl.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": fields})
    return err
}


