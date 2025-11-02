package battle

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastDestroyTheLight reduces warrior attack and defense by 30% (stackable up to 2 times: 70% → 49%)
func (s *Service) CastDestroyTheLight(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) error {
	// Get battle
	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return errors.New("battle not found")
	}

	if battle.Status != BattleStatusInProgress {
		return errors.New("battle must be in progress to cast spell")
	}

	// Check existing Destroy the Light spells (stackable up to 2 times)
	var existingSpells []Spell
	cursor, err := SpellColl.Find(ctx, bson.M{
		"battle_id":  battleID,
		"spell_type": SpellDestroyTheLight,
		"is_active": true,
	})
	if err == nil {
		defer cursor.Close(ctx)
		cursor.All(ctx, &existingSpells)
	}

	currentStackCount := len(existingSpells)
	if currentStackCount >= 2 {
		return errors.New("Destroy the Light spell can only be stacked 2 times (already at max)")
	}

	// Calculate reduction multiplier
	// First cast: 100% → 70% (0.7 multiplier)
	// Second cast: 70% → 49% (0.7 * 0.7 = 0.49 multiplier)
	var reductionMultiplier float64
	if currentStackCount == 0 {
		reductionMultiplier = 0.7 // First stack: 30% reduction
	} else {
		reductionMultiplier = 0.49 // Second stack: 51% total reduction (49% remaining)
	}

	// Get all warrior participants on light side
	filter := bson.M{
		"battle_id": battleID,
		"type":      ParticipantTypeWarrior,
		"side":      TeamSideLight,
		"is_alive":  true,
	}

	cursor, err = BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find warriors: %w", err)
	}
	defer cursor.Close(ctx)

	var warriors []BattleParticipant
	if err := cursor.All(ctx, &warriors); err != nil {
		return fmt.Errorf("failed to decode warriors: %w", err)
	}

	// Reduce attack and defense for all warriors
	updatedCount := 0
	for _, warrior := range warriors {
		newAttackPower := int(float64(warrior.AttackPower) * reductionMultiplier)
		newDefense := int(float64(warrior.Defense) * reductionMultiplier)

		// Minimum values (at least 1)
		if newAttackPower < 1 {
			newAttackPower = 1
		}
		if newDefense < 1 {
			newDefense = 1
		}

		updateData := bson.M{
			"attack_power": newAttackPower,
			"defense":      newDefense,
			"updated_at":   battle.UpdatedAt,
		}

		_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": warrior.ID}, bson.M{"$set": updateData})
		if err != nil {
			log.Printf("Failed to reduce warrior %s stats: %v", warrior.Name, err)
			continue
		}

		updatedCount++
	}

	// Create spell record
	newStackCount := currentStackCount + 1
	spell := &Spell{
		BattleID:      battleID,
		SpellType:     SpellDestroyTheLight,
		Side:          TeamSideDark,
		CasterUsername: casterUsername,
		CasterUserID:   casterUserID,
		CasterRole:    "dark_king",
		StackCount:    newStackCount,
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
		remainingPercent := int(reductionMultiplier * 100)
		message := fmt.Sprintf("💀 SPELL CAST: Destroy the Light! Warrior stats reduced to %d%% (Stack: %d/2)! (%d warriors affected)",
			remainingPercent, newStackCount, updatedCount)
		if err := LogBattleEvent(ctx, battleID, "spell_cast", message); err != nil {
			log.Printf("Failed to log spell cast: %v", err)
		}
	}()

	log.Printf("Destroy the Light spell cast by %s in battle %s - Stack %d/2, %d warriors affected", casterUsername, battleID.Hex(), newStackCount, updatedCount)
	return nil
}

