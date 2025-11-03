package battle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"network-sec-micro/internal/battle/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StartArenaBattle starts a 1v1 arena battle between two warriors
func (s *Service) StartArenaBattle(ctx context.Context, cmd dto.StartArenaBattleCommand) (*Battle, []*BattleParticipant, error) {
	// Get both warriors via gRPC
	player1, err := GetWarriorByID(ctx, cmd.Player1ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get player1 info: %w", err)
	}

	player2, err := GetWarriorByID(ctx, cmd.Player2ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get player2 info: %w", err)
	}

	// Calculate HP based on total power
	player1MaxHP := int(player1.TotalPower) * 10
	if player1MaxHP < 100 {
		player1MaxHP = 100
	}

	player2MaxHP := int(player2.TotalPower) * 10
	if player2MaxHP < 100 {
		player2MaxHP = 100
	}

	// Determine sides (random or based on some logic - for arena, we can assign light/dark randomly)
	// For simplicity, player1 = light, player2 = dark
	lightSideName := fmt.Sprintf("%s (Light)", player1.Username)
	darkSideName := fmt.Sprintf("%s (Dark)", player2.Username)

	maxTurns := cmd.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 50 // Default for arena battles
	}

	// Create battle
	now := time.Now()
	battle := &Battle{
		BattleType:            BattleTypeTeam,
		LightSideName:         lightSideName,
		DarkSideName:          darkSideName,
		CurrentTurn:           0,
		CurrentParticipantIndex: 0,
		MaxTurns:              maxTurns,
		Status:                BattleStatusPending,
		CreatedBy:             "arena_system",
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	result, err := BattleColl.InsertOne(ctx, battle)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create battle: %w", err)
	}

	battle.ID = result.InsertedID.(primitive.ObjectID)

	// Create participants
	participants := []*BattleParticipant{
		{
			BattleID:      battle.ID,
			ParticipantID: fmt.Sprintf("%d", player1.Id),
			Name:          player1.Username,
			Type:          ParticipantTypeWarrior,
			Side:          TeamSideLight,
			HP:            player1MaxHP,
			MaxHP:         player1MaxHP,
			AttackPower:   int(player1.AttackPower),
			Defense:       int(player1.Defense),
			IsAlive:       true,
			IsDefeated:    false,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			BattleID:      battle.ID,
			ParticipantID: fmt.Sprintf("%d", player2.Id),
			Name:          player2.Username,
			Type:          ParticipantTypeWarrior,
			Side:          TeamSideDark,
			HP:            player2MaxHP,
			MaxHP:         player2MaxHP,
			AttackPower:   int(player2.AttackPower),
			Defense:       int(player2.Defense),
			IsAlive:       true,
			IsDefeated:    false,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}

	// Insert participants
	if len(participants) > 0 {
		docs := make([]interface{}, len(participants))
		for i, p := range participants {
			docs[i] = p
		}
		_, err = BattleParticipantColl.InsertMany(ctx, docs)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create participants: %w", err)
		}
	}

	// Start the battle
	battle.Status = BattleStatusInProgress
	battle.StartedAt = &now

	updateData := bson.M{
		"status":     battle.Status,
		"started_at": battle.StartedAt,
		"updated_at": time.Now(),
	}

	_, err = BattleColl.UpdateOne(ctx, bson.M{"_id": battle.ID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start battle: %w", err)
	}

	// Publish battle started event
	go PublishBattleStartedEvent(
		battle.ID.Hex(),
		string(battle.BattleType),
		cmd.Player1ID,
		player1.Username,
		fmt.Sprintf("%d", player2.Id),
		player2.Username,
		"warrior",
	)

	return battle, participants, nil
}

