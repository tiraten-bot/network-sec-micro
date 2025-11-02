package battle

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastCallOfTheLightKing doubles attack power for all warrior units for the entire battle duration
func (s *Service) CastCallOfTheLightKing(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) error {
	// Get battle
	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return errors.New("battle not found")
	}

	if battle.Status != BattleStatusInProgress {
		return errors.New("battle must be in progress to cast spell")
	}

	// Check if spell already cast (should be unique per battle)
	var existingSpell Spell
	err = SpellColl.FindOne(ctx, bson.M{
		"battle_id":  battleID,
		"spell_type": SpellCallOfTheLightKing,
		"is_active": true,
	}).Decode(&existingSpell)

	if err == nil {
		return errors.New("Call of the Light King spell is already active in this battle")
	}

	// Get all warrior participants on light side
	filter := bson.M{
		"battle_id": battleID,
		"type":      ParticipantTypeWarrior,
		"side":      TeamSideLight,
		"is_alive":  true,
	}

	cursor, err := BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find warriors: %w", err)
	}
	defer cursor.Close(ctx)

	var warriors []BattleParticipant
	if err := cursor.All(ctx, &warriors); err != nil {
		return fmt.Errorf("failed to decode warriors: %w", err)
	}

	// Double attack power for all warriors
	updatedCount := 0
	for _, warrior := range warriors {
		newAttackPower := warrior.AttackPower * 2

		updateData := bson.M{
			"attack_power": newAttackPower,
			"updated_at":   battle.UpdatedAt,
		}

		_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": warrior.ID}, bson.M{"$set": updateData})
		if err != nil {
			log.Printf("Failed to update warrior %s attack power: %v", warrior.Name, err)
			continue
		}

		updatedCount++
	}

	// Create spell record
	spell := &Spell{
		BattleID:      battleID,
		SpellType:     SpellCallOfTheLightKing,
		Side:          TeamSideLight,
		CasterUsername: casterUsername,
		CasterUserID:   casterUserID,
		CasterRole:    "light_king",
		IsActive:      true,
		CastAt:        battle.UpdatedAt,
		CreatedAt:     battle.UpdatedAt,
		UpdatedAt:     battle.UpdatedAt,
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("✨ SPELL CAST: Call of the Light King! All warrior attack power doubled! (%d warriors affected)", updatedCount)
		if err := LogBattleEvent(ctx, battleID, "spell_cast", message); err != nil {
			log.Printf("Failed to log spell cast: %v", err)
		}
	}()

	log.Printf("Call of the Light King spell cast by %s in battle %s - %d warriors affected", casterUsername, battleID.Hex(), updatedCount)
	return nil
}

