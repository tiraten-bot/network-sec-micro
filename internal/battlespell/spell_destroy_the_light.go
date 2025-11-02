package battlespell

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pbBattle "network-sec-micro/api/proto/battle"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastDestroyTheLight reduces warrior attack and defense by 30% (stackable up to 2 times: 70% → 49%)
func (s *Service) CastDestroyTheLight(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) (int, error) {
	// Get battle via gRPC
	battleIDStr := battleID.Hex()
	battle, err := GetBattleByID(ctx, battleIDStr)
	if err != nil {
		return 0, errors.New("battle not found")
	}

	if battle.Status != "in_progress" {
		return 0, errors.New("battle must be in progress to cast spell")
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
		return 0, errors.New("Destroy the Light spell can only be stacked 2 times (already at max)")
	}

	// Calculate reduction multiplier
	var reductionMultiplier float64
	if currentStackCount == 0 {
		reductionMultiplier = 0.7 // First stack: 30% reduction
	} else {
		reductionMultiplier = 0.49 // Second stack: 51% total reduction (49% remaining)
	}

	// Get all warrior participants on light side via gRPC
	participants, err := GetBattleParticipants(ctx, battleIDStr, "light")
	if err != nil {
		return 0, fmt.Errorf("failed to get battle participants: %w", err)
	}

	// Reduce attack and defense for all warriors
	updatedCount := 0
	for _, p := range participants {
		if p.Type == "warrior" && p.IsAlive {
			newAttackPower := int32(float64(p.AttackPower) * reductionMultiplier)
			newDefense := int32(float64(p.Defense) * reductionMultiplier)

			// Minimum values (at least 1)
			if newAttackPower < 1 {
				newAttackPower = 1
			}
			if newDefense < 1 {
				newDefense = 1
			}

			err = UpdateParticipantStats(ctx, battleIDStr, p.ParticipantId, p.Hp, p.MaxHp, newAttackPower, newDefense, p.IsAlive)
			if err != nil {
				log.Printf("Failed to reduce warrior %s stats: %v", p.Name, err)
				continue
			}

			updatedCount++
		}
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
		CastAt:        time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	log.Printf("Destroy the Light spell cast by %s in battle %s - Stack %d/2, %d warriors affected", casterUsername, battleID.Hex(), newStackCount, updatedCount)
	return updatedCount, nil
}
