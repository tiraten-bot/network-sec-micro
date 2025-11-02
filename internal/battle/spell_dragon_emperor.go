package battle

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CastDragonEmperor adds Dark Emperor's stats to dragon for the entire battle duration
func (s *Service) CastDragonEmperor(ctx context.Context, battleID primitive.ObjectID, dragonParticipantID string, darkEmperorParticipantID string, casterUsername string, casterUserID string) error {
	// Get battle
	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		return errors.New("battle not found")
	}

	if battle.Status != BattleStatusInProgress {
		return errors.New("battle must be in progress to cast spell")
	}

	// Get dragon participant
	var dragonParticipant BattleParticipant
	err = BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": dragonParticipantID,
		"type":          ParticipantTypeDragon,
	}).Decode(&dragonParticipant)

	if err != nil {
		return errors.New("dragon participant not found")
	}

	// Get Dark Emperor participant
	var darkEmperorParticipant BattleParticipant
	err = BattleParticipantColl.FindOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": darkEmperorParticipantID,
		"type":          ParticipantTypeDarkEmperor,
	}).Decode(&darkEmperorParticipant)

	if err != nil {
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
	originalAttack := dragonParticipant.AttackPower
	originalDefense := dragonParticipant.Defense
	originalMaxHP := dragonParticipant.MaxHP
	originalHP := dragonParticipant.HP

	newAttackPower := dragonParticipant.AttackPower + darkEmperorParticipant.AttackPower
	newDefense := dragonParticipant.Defense + darkEmperorParticipant.Defense
	newMaxHP := dragonParticipant.MaxHP + darkEmperorParticipant.MaxHP
	// Add HP bonus proportionally
	hpBonus := darkEmperorParticipant.MaxHP
	newHP := dragonParticipant.HP + hpBonus
	if newHP > newMaxHP {
		newHP = newMaxHP
	}

	updateData := bson.M{
		"attack_power": newAttackPower,
		"defense":      newDefense,
		"max_hp":       newMaxHP,
		"hp":           newHP,
		"updated_at":   battle.UpdatedAt,
	}

	_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{"_id": dragonParticipant.ID}, bson.M{"$set": updateData})
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
		CastAt:             battle.UpdatedAt,
		CreatedAt:          battle.UpdatedAt,
		UpdatedAt:          battle.UpdatedAt,
	}

	_, err = SpellColl.InsertOne(ctx, spell)
	if err != nil {
		log.Printf("Warning: failed to record spell cast: %v", err)
	}

	// Log to Redis
	go func() {
		message := fmt.Sprintf("🐉 SPELL CAST: Dragon Emperor! %s enhanced by Dark Emperor stats! (Attack: %d→%d, Defense: %d→%d, HP: %d→%d)",
			dragonParticipant.Name, originalAttack, newAttackPower, originalDefense, newDefense, originalMaxHP, newMaxHP)
		if err := LogBattleEvent(ctx, battleID, "spell_cast", message); err != nil {
			log.Printf("Failed to log spell cast: %v", err)
		}
	}()

	log.Printf("Dragon Emperor spell cast by %s in battle %s - Dragon %s enhanced", casterUsername, battleID.Hex(), dragonParticipant.Name)
	return nil
}

