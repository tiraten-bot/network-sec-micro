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

// CastCallOfTheLightKing doubles attack power for all warrior units for the entire battle duration
func (s *Service) CastCallOfTheLightKing(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) (int, error) {
	// Get battle via gRPC
	battleIDStr := battleID.Hex()
	battle, err := GetBattleByID(ctx, battleIDStr)
	if err != nil {
		return 0, errors.New("battle not found")
	}

	if battle.Status != "in_progress" {
		return 0, errors.New("battle must be in progress to cast spell")
	}

	// Check if spell already cast (should be unique per battle)
	var existingSpell Spell
	err = SpellColl.FindOne(ctx, bson.M{
		"battle_id":  battleID,
		"spell_type": SpellCallOfTheLightKing,
		"is_active": true,
	}).Decode(&existingSpell)

	if err == nil {
		return 0, errors.New("Call of the Light King spell is already active in this battle")
	}

	// Get all warrior participants on light side via gRPC
	participants, err := GetBattleParticipants(ctx, battleIDStr, "light")
	if err != nil {
		return 0, fmt.Errorf("failed to get battle participants: %w", err)
	}

	// Filter warriors and double attack power
	updatedCount := 0
	for _, p := range participants {
		if p.Type == "warrior" && p.IsAlive {
			newAttackPower := p.AttackPower * 2

			err = UpdateParticipantStats(ctx, battleIDStr, p.ParticipantId, p.Hp, p.MaxHp, newAttackPower, p.Defense, p.IsAlive)
			if err != nil {
				log.Printf("Failed to update warrior %s attack power: %v", p.Name, err)
				continue
			}

			updatedCount++
		}
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
		CastAt:        time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	log.Printf("Call of the Light King spell cast by %s in battle %s - %d warriors affected", casterUsername, battleID.Hex(), updatedCount)
	return updatedCount, nil
}

