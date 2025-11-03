package arenaspell

import (
    "context"
    "errors"
    "fmt"
    "time"

    "network-sec-micro/internal/arenaspell/dto"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Service provides arenaspell business logic (1v1 oriented)
type Service struct{}

func NewService() *Service { return &Service{} }

// CastSpell applies a 1v1 spell effect for a given match and persists spell state
// Returns affected count (1 or 2 depending on effect) for feedback
func (s *Service) CastSpell(ctx context.Context, cmd dto.CastArenaSpellCommand) (int, error) {
    st := SpellType(cmd.SpellType)
    if st != SpellCallOfTheLightKing && st != SpellResistance && st != SpellRebirth && st != SpellDestroyTheLight {
        return 0, errors.New("invalid spell type")
    }

    matchID, err := primitive.ObjectIDFromHex(cmd.MatchID)
    if err != nil {
        return 0, errors.New("invalid match ID")
    }

    // Persist spell cast (upsert stack logic for destroy_the_light)
    now := time.Now()

    if st == SpellDestroyTheLight {
        // Increase stack up to 2 for given match and caster
        filter := bson.M{"match_id": matchID, "spell_type": st, "is_active": true}
        var existing Spell
        err := SpellColl.FindOne(ctx, filter).Decode(&existing)
        if err == nil {
            if existing.StackCount >= 2 {
                return 0, errors.New("destroy_the_light already at max stacks (2)")
            }
            update := bson.M{"$set": bson.M{"updated_at": now}, "$inc": bson.M{"stack_count": 1}}
            _, _ = SpellColl.UpdateOne(ctx, filter, update)
        } else {
            rec := &Spell{
                MatchID:        matchID,
                SpellType:      st,
                CasterUserID:   cmd.CasterUserID,
                CasterUsername: cmd.CasterUsername,
                StackCount:     1,
                IsActive:       true,
                CastAt:         now,
                CreatedAt:      now,
                UpdatedAt:      now,
            }
            _, _ = SpellColl.InsertOne(ctx, rec)
        }
    } else {
        rec := &Spell{
            MatchID:        matchID,
            SpellType:      st,
            CasterUserID:   cmd.CasterUserID,
            CasterUsername: cmd.CasterUsername,
            IsActive:       true,
            CastAt:         now,
            CreatedAt:      now,
            UpdatedAt:      now,
        }
        _, _ = SpellColl.InsertOne(ctx, rec)
    }

    // Apply effect via Arena service (to be wired through gRPC in next step)
    // For now, we only return affected count and assume external application in Arena turn loop.
    switch st {
    case SpellCallOfTheLightKing:
        return 1, nil // caster buff
    case SpellResistance:
        return 1, nil // caster buff
    case SpellRebirth:
        return 1, nil // caster revive capability
    case SpellDestroyTheLight:
        return 1, nil // opponent debuff stack incremented
    default:
        return 0, fmt.Errorf("spell not implemented: %s", st)
    }
}


