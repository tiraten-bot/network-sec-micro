package battle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastRebirth revives all defeated warrior units
func (s *Service) CastRebirth(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) (int, error) {
	// Get battle
	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return 0, errors.New("battle not found")
	}

	if battle.Status != BattleStatusInProgress {
		return 0, errors.New("battle must be in progress to cast spell")
	}

	// Get all defeated warrior participants on light side
	filter := bson.M{
		"battle_id":   battleID,
		"type":        ParticipantTypeWarrior,
		"side":        TeamSideLight,
		"is_defeated": true,
	}

	cursor, err := BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to find defeated warriors: %w", err)
	}
	defer cursor.Close(ctx)

	var defeatedWarriors []BattleParticipant
	if err := cursor.All(ctx, &defeatedWarriors); err != nil {
		return 0, fmt.Errorf("failed to decode warriors: %w", err)
	}

	if len(defeatedWarriors) == 0 {
		return 0, errors.New("no defeated warriors to revive")
	}

	// Revive all defeated warriors
	revivedCount := 0
	for _, warrior := range defeatedWarriors {
		warrior.HP = warrior.MaxHP
		warrior.IsAlive = true
		warrior.IsDefeated = false
		warrior.DefeatedAt = nil
		warrior.UpdatedAt = time.Now()

		updateData := bson.M{
			"hp":          warrior.HP,
			"is_alive":    warrior.IsAlive,
			"is_defeated": warrior.IsDefeated,
			"defeated_at": nil,
			"updated_at":  warrior.UpdatedAt,
		}

		_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": warrior.ID}, bson.M{"$set": updateData})
		if err != nil {
			log.Printf("Failed to revive warrior %s: %v", warrior.Name, err)
			continue
		}

		revivedCount++
	}

	// Create spell record
	spell := &Spell{
		BattleID:      battleID,
		SpellType:     SpellRebirth,
		Side:          TeamSideLight,
		CasterUsername: casterUsername,
		CasterUserID:   casterUserID,
		CasterRole:    "light_king",
		IsActive:      false, // Rebirth is instant, not a lasting effect
		CastAt:        time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("🌟 SPELL CAST: Rebirth! All defeated warriors revived! (%d warriors revived)", revivedCount)
		if err := LogBattleEvent(ctx, battleID, "spell_cast", message); err != nil {
			log.Printf("Failed to log spell cast: %v", err)
		}
	}()

	log.Printf("Rebirth spell cast by %s in battle %s - %d warriors revived", casterUsername, battleID.Hex(), revivedCount)
	return revivedCount, nil
}

