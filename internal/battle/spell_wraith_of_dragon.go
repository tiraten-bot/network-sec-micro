package battle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastWraithOfDragon enables wraith effect: when dragon kills warrior, random warrior also dies (max 25 times)
func (s *Service) CastWraithOfDragon(ctx context.Context, battleID primitive.ObjectID, casterUsername string, casterUserID string) error {
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
		CastAt:        battle.UpdatedAt,
		CreatedAt:     battle.UpdatedAt,
		UpdatedAt:     battle.UpdatedAt,
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		return fmt.Errorf("failed to record spell cast: %w", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("👻 SPELL CAST: Wraith of Dragon! When dragon kills a warrior, a random warrior will also die! (Max 25 times)")
		if err := LogBattleEvent(ctx, battleID, "spell_cast", message); err != nil {
			log.Printf("Failed to log spell cast: %v", err)
		}
	}()

	log.Printf("Wraith of Dragon spell cast by %s in battle %s", casterUsername, battleID.Hex())
	return nil
}

// GetBattle retrieves battle - helper for wraith trigger
func getBattleForSpell(ctx context.Context, battleID primitive.ObjectID) (*Battle, error) {
	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return nil, err
	}
	return &battle, nil
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

	// Get battle for timestamp
	battle, err := getBattleForSpell(ctx, battleID)
	if err != nil {
		return "", fmt.Errorf("failed to get battle: %w", err)
	}

	// Check if max wraith count reached
	if spell.WraithCount >= 25 {
		// Deactivate spell
		_, err = SpellColl.UpdateOne(ctx, bson.M{"_id": spell.ID}, bson.M{"$set": bson.M{
			"is_active": false,
			"updated_at": battle.UpdatedAt,
		}})
		return "", nil
	}

	// Get all alive warrior participants on light side
	filter := bson.M{
		"battle_id": battleID,
		"type":      ParticipantTypeWarrior,
		"side":      TeamSideLight,
		"is_alive":  true,
	}

	cursor, err := BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("failed to find warriors: %w", err)
	}
	defer cursor.Close(ctx)

	var aliveWarriors []BattleParticipant
	if err := cursor.All(ctx, &aliveWarriors); err != nil {
		return "", fmt.Errorf("failed to decode warriors: %w", err)
	}

	if len(aliveWarriors) == 0 {
		// No warriors left to destroy
		return "", nil
	}

	// Select random warrior to destroy
	randomIndex := rand.Intn(len(aliveWarriors))
	targetWarrior := aliveWarriors[randomIndex]

	// Destroy the random warrior
	targetWarrior.HP = 0
	targetWarrior.IsAlive = false
	targetWarrior.IsDefeated = true
	now := time.Now()
	targetWarrior.DefeatedAt = &now
	targetWarrior.UpdatedAt = battle.UpdatedAt

	updateData := bson.M{
		"hp":          targetWarrior.HP,
		"is_alive":    false,
		"is_defeated": true,
		"defeated_at": targetWarrior.DefeatedAt,
		"updated_at":  targetWarrior.UpdatedAt,
	}

	_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": targetWarrior.ID}, bson.M{"$set": updateData})
	if err != nil {
		return "", fmt.Errorf("failed to destroy warrior: %w", err)
	}

	// Increment wraith count
	newWraithCount := spell.WraithCount + 1
	_, err = SpellColl.UpdateOne(ctx, bson.M{"_id": spell.ID}, bson.M{"$set": bson.M{
		"wraith_count": newWraithCount,
		"updated_at":   battle.UpdatedAt,
	}})
	if err != nil {
		log.Printf("Warning: failed to update wraith count: %v", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("👻 WRAITH TRIGGERED: %s destroyed by Wraith of Dragon! (Wraith count: %d/25)", targetWarrior.Name, newWraithCount)
		if err := LogBattleEvent(ctx, battleID, "wraith_triggered", message); err != nil {
			log.Printf("Failed to log wraith trigger: %v", err)
		}
	}()

	log.Printf("Wraith of Dragon triggered in battle %s - %s destroyed (count: %d/25)", battleID.Hex(), targetWarrior.Name, newWraithCount)
	return targetWarrior.ParticipantID, nil
}

