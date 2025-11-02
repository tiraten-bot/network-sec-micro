package battlespell

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	pbBattle "network-sec-micro/api/proto/battle"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastWraithOfDragon enables wraith effect: when dragon kills warrior, random warrior also dies (max 25 times)
func (s *Service) CastWraithOfDragon(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) error {
	// Get battle via gRPC
	battleIDStr := battleID.Hex()
	battle, err := GetBattleByID(ctx, battleIDStr)
	if err != nil {
		return errors.New("battle not found")
	}

	if battle.Status != "in_progress" {
		return errors.New("battle must be in progress to cast spell")
	}

	// Check if spell already cast
	var existingSpell Spell
	err = SpellColl.FindOne(ctx, bson.M{
		"battle_id":  battleID,
		"spell_type": SpellWraithOfDragon,
		"is_active": true,
	}).Decode(&existingSpell)

	if err == nil {
		return errors.New("Wraith of Dragon spell is already active in this battle")
	}

	// Create spell record
	spell := &Spell{
		BattleID:      battleID,
		SpellType:     SpellWraithOfDragon,
		Side:          TeamSideDark,
		CasterUsername: casterUsername,
		CasterUserID:   casterUserID,
		CasterRole:    "dark_king",
		WraithCount:   0, // Start at 0, max 25
		IsActive:      true,
		CastAt:        time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		return fmt.Errorf("failed to record spell cast: %w", err)
	}

	log.Printf("Wraith of Dragon spell cast by %s in battle %s", casterUsername, battleID.Hex())
	return nil
}

// TriggerWraithOfDragon triggers the wraith effect when dragon kills a warrior
// Returns the additional warrior ID that was destroyed, or empty string if none
func (s *Service) TriggerWraithOfDragon(ctx context.Context, battleID primitive.ObjectID) (string, error) {
	// Get active Wraith of Dragon spell
	var spell Spell
	err := SpellColl.FindOne(ctx, bson.M{
		"battle_id":  battleID,
		"spell_type": SpellWraithOfDragon,
		"is_active": true,
	}).Decode(&spell)

	if err != nil {
		// Spell not active, nothing to do
		return "", nil
	}

	// Check if max wraith count reached
	if spell.WraithCount >= 25 {
		// Deactivate spell
		_, err = SpellColl.UpdateOne(ctx, bson.M{"_id": spell.ID}, bson.M{"$set": bson.M{
			"is_active": false,
			"updated_at": time.Now(),
		}})
		return "", nil
	}

	// Get all alive warrior participants on light side via gRPC
	battleIDStr := battleID.Hex()
	participants, err := GetBattleParticipants(ctx, battleIDStr, "light")
	if err != nil {
		return "", fmt.Errorf("failed to get battle participants: %w", err)
	}

	// Filter alive warriors
	var aliveWarriors []*pbBattle.BattleParticipant
	for _, p := range participants {
		if p.Type == "warrior" && p.IsAlive {
			aliveWarriors = append(aliveWarriors, p)
		}
	}

	if len(aliveWarriors) == 0 {
		return "", nil
	}

	// Select random warrior to destroy
	randomIndex := rand.Intn(len(aliveWarriors))
	targetWarrior := aliveWarriors[randomIndex]

	// Destroy the random warrior via gRPC
	err = UpdateParticipantStats(ctx, battleIDStr, targetWarrior.ParticipantId, 0, targetWarrior.MaxHp, targetWarrior.AttackPower, targetWarrior.Defense, false)
	if err != nil {
		return "", fmt.Errorf("failed to destroy warrior: %w", err)
	}

	// Increment wraith count
	newWraithCount := spell.WraithCount + 1
	_, err = SpellColl.UpdateOne(ctx, bson.M{"_id": spell.ID}, bson.M{"$set": bson.M{
		"wraith_count": newWraithCount,
		"updated_at":   time.Now(),
	}})
	if err != nil {
		log.Printf("Warning: failed to update wraith count: %v", err)
	}

	log.Printf("Wraith of Dragon triggered in battle %s - %s destroyed (count: %d/25)", battleID.Hex(), targetWarrior.Name, newWraithCount)
	return targetWarrior.ParticipantId, nil
}
