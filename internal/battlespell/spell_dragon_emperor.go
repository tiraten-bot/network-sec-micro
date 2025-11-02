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

// CastDragonEmperor adds Dark Emperor's stats to dragon for the entire battle duration
func (s *Service) CastDragonEmperor(ctx context.Context, battleID primitive.ObjectID, dragonParticipantID string, darkEmperorParticipantID string, casterUsername string, casterUserID string) error {
	// Get battle via gRPC
	battleIDStr := battleID.Hex()
	battle, err := GetBattleByID(ctx, battleIDStr)
	if err != nil {
		return errors.New("battle not found")
	}

	if battle.Status != "in_progress" {
		return errors.New("battle must be in progress to cast spell")
	}

	// Get all participants via gRPC
	participants, err := GetBattleParticipants(ctx, battleIDStr, "")
	if err != nil {
		return fmt.Errorf("failed to get battle participants: %w", err)
	}

	// Find dragon and dark emperor participants
	var dragonParticipant *pbBattle.BattleParticipant
	var darkEmperorParticipant *pbBattle.BattleParticipant

	for _, p := range participants {
		if p.ParticipantId == dragonParticipantID && p.Type == "dragon" {
			dragonParticipant = p
		}
		if p.ParticipantId == darkEmperorParticipantID && p.Type == "dark_emperor" {
			darkEmperorParticipant = p
		}
	}

	if dragonParticipant == nil {
		return errors.New("dragon participant not found")
	}
	if darkEmperorParticipant == nil {
		return errors.New("dark emperor participant not found in battle")
	}

	// Check if spell already cast for this dragon
	var existingSpell Spell
	err = SpellColl.FindOne(ctx, bson.M{
		"battle_id":        battleID,
		"spell_type":       SpellDragonEmperor,
		"target_dragon_id": dragonParticipantID,
		"is_active":        true,
	}).Decode(&existingSpell)

	if err == nil {
		return errors.New("Dragon Emperor spell is already active for this dragon")
	}

	// Add Dark Emperor stats to dragon
	newAttackPower := dragonParticipant.AttackPower + darkEmperorParticipant.AttackPower
	newDefense := dragonParticipant.Defense + darkEmperorParticipant.Defense
	newMaxHP := dragonParticipant.MaxHp + darkEmperorParticipant.MaxHp
	hpBonus := darkEmperorParticipant.MaxHp
	newHP := dragonParticipant.Hp + hpBonus
	if newHP > newMaxHP {
		newHP = newMaxHP
	}

	err = UpdateParticipantStats(ctx, battleIDStr, dragonParticipantID, int32(newHP), int32(newMaxHP), newAttackPower, newDefense, dragonParticipant.IsAlive)
	if err != nil {
		return fmt.Errorf("failed to enhance dragon: %w", err)
	}

	// Create spell record
	spell := &Spell{
		BattleID:           battleID,
		SpellType:          SpellDragonEmperor,
		Side:               TeamSideDark,
		CasterUsername:     casterUsername,
		CasterUserID:       casterUserID,
		CasterRole:         "dark_king",
		TargetDragonID:     dragonParticipantID,
		TargetDarkEmperorID: darkEmperorParticipantID,
		IsActive:           true,
		CastAt:             time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	log.Printf("Dragon Emperor spell cast by %s in battle %s - Dragon enhanced", casterUsername, battleID.Hex())
	return nil
}
