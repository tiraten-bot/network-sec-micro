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

// CastRebirth revives all defeated warrior units
func (s *Service) CastRebirth(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) (int, error) {
	// Get battle via gRPC
	battleIDStr := battleID.Hex()
	battle, err := GetBattleByID(ctx, battleIDStr)
	if err != nil {
		return 0, errors.New("battle not found")
	}

	if battle.Status != "in_progress" {
		return 0, errors.New("battle must be in progress to cast spell")
	}

	// Get all participants on light side via gRPC
	participants, err := GetBattleParticipants(ctx, battleIDStr, "light")
	if err != nil {
		return 0, fmt.Errorf("failed to get battle participants: %w", err)
	}

	// Filter defeated warriors
	var defeatedWarriors []*pbBattle.BattleParticipant
	for _, p := range participants {
		if p.Type == "warrior" && p.IsDefeated {
			defeatedWarriors = append(defeatedWarriors, p)
		}
	}

	if len(defeatedWarriors) == 0 {
		return 0, errors.New("no defeated warriors to revive")
	}

	// Revive all defeated warriors
	revivedCount := 0
	for _, p := range defeatedWarriors {
		err = UpdateParticipantStats(ctx, battleIDStr, p.ParticipantId, p.MaxHp, p.MaxHp, p.AttackPower, p.Defense, true)
		if err != nil {
			log.Printf("Failed to revive warrior %s: %v", p.Name, err)
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

	log.Printf("Rebirth spell cast by %s in battle %s - %d warriors revived", casterUsername, battleID.Hex(), revivedCount)
	return revivedCount, nil
}
