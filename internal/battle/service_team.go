package battle

import (
    "context"
    "errors"
    "fmt"
    "log"
    "math/rand"
    "time"

    "network-sec-micro/internal/battle/dto"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// StartBattle creates and starts a new team-based battle
func (s *Service) StartBattle(cmd dto.StartBattleCommand) (*Battle, []*BattleParticipant, error) {
	ctx := context.Background()

	// Validate that we have participants on both sides
	if len(cmd.LightParticipants) == 0 {
		return nil, nil, errors.New("light side must have at least one participant")
	}
	if len(cmd.DarkParticipants) == 0 {
		return nil, nil, errors.New("dark side must have at least one participant")
	}

	// Validate team composition rules
	if err := ValidateBattleParticipants(cmd); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if any warrior participants are currently healing
	for _, p := range cmd.LightParticipants {
		if p.Type == "warrior" {
			// Parse warrior ID from participant_id
			var warriorID uint
			if _, err := fmt.Sscanf(p.ParticipantID, "%d", &warriorID); err == nil {
				if err := CheckWarriorCanBattle(ctx, warriorID); err != nil {
					return nil, nil, fmt.Errorf("participant %s cannot battle: %w", p.Name, err)
				}
			}
		}
	}
	for _, p := range cmd.DarkParticipants {
		if p.Type == "warrior" {
			var warriorID uint
			if _, err := fmt.Sscanf(p.ParticipantID, "%d", &warriorID); err == nil {
				if err := CheckWarriorCanBattle(ctx, warriorID); err != nil {
					return nil, nil, fmt.Errorf("participant %s cannot battle: %w", p.Name, err)
				}
			}
		}
	}

	// Set defaults
	lightSideName := cmd.LightSideName
	if lightSideName == "" {
		lightSideName = "Light Alliance"
	}
	darkSideName := cmd.DarkSideName
	if darkSideName == "" {
		darkSideName = "Dark Forces"
	}

	maxTurns := cmd.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 100 // Default for team battles
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
		CreatedBy:             cmd.CreatedBy,
		CreatedAt:             now,
		UpdatedAt:             now,
        WagerAmount:           cmd.WagerAmount,
        LightEmperorID:        cmd.LightEmperorID,
        DarkEmperorID:         cmd.DarkEmperorID,
	}

    battleID, err := GetRepository().CreateBattle(ctx, battle)
    if err != nil { return nil, nil, fmt.Errorf("failed to create battle: %w", err) }
    battle.ID = battleID

	// Create participants
	participants := make([]*BattleParticipant, 0, len(cmd.LightParticipants)+len(cmd.DarkParticipants))

	// Add light participants
	for _, pInfo := range cmd.LightParticipants {
		participant := &BattleParticipant{
            BattleID:      battle.ID,
			ParticipantID: pInfo.ParticipantID,
			Name:         pInfo.Name,
			Type:         ParticipantType(pInfo.Type),
			Side:         TeamSideLight,
			HP:           pInfo.HP,
			MaxHP:        pInfo.MaxHP,
			AttackPower:  pInfo.AttackPower,
			Defense:      pInfo.Defense,
			IsAlive:      true,
			IsDefeated:   false,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// Set defaults if not provided
		if participant.HP <= 0 {
			participant.HP = participant.MaxHP
		}
		if participant.MaxHP <= 0 {
			participant.MaxHP = participant.HP
		}
		if participant.HP == 0 && participant.MaxHP == 0 {
			participant.MaxHP = 100 // Default
			participant.HP = participant.MaxHP
		}

		participants = append(participants, participant)
	}

	// Add dark participants
	for _, pInfo := range cmd.DarkParticipants {
		participant := &BattleParticipant{
			BattleID:      battle.ID,
			ParticipantID: pInfo.ParticipantID,
			Name:         pInfo.Name,
			Type:         ParticipantType(pInfo.Type),
			Side:         TeamSideDark,
			HP:           pInfo.HP,
			MaxHP:        pInfo.MaxHP,
			AttackPower:  pInfo.AttackPower,
			Defense:      pInfo.Defense,
			IsAlive:      true,
			IsDefeated:   false,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// Set defaults if not provided
		if participant.HP <= 0 {
			participant.HP = participant.MaxHP
		}
		if participant.MaxHP <= 0 {
			participant.MaxHP = participant.HP
		}
		if participant.HP == 0 && participant.MaxHP == 0 {
			participant.MaxHP = 100 // Default
			participant.HP = participant.MaxHP
		}

		participants = append(participants, participant)
	}

	// Insert all participants
    if len(participants) > 0 {
        if err := GetRepository().InsertParticipants(ctx, participants); err != nil {
            return nil, nil, fmt.Errorf("failed to create participants: %w", err)
        }
    }

    // Start the battle if no emperor approval required or wager is zero
    if !(cmd.RequireEmperorApproval && battle.WagerAmount > 0) {
        battle.Status = BattleStatusInProgress
        battle.StartedAt = &now
    }

    updateData := map[string]interface{}{
        "status":     battle.Status,
        "started_at": battle.StartedAt,
        "updated_at": time.Now(),
        "wager_amount": battle.WagerAmount,
        "light_emperor_id": battle.LightEmperorID,
        "dark_emperor_id": battle.DarkEmperorID,
    }
    if err := GetRepository().UpdateBattleFields(ctx, battle.ID, updateData); err != nil { return nil, nil, fmt.Errorf("failed to update battle: %w", err) }

	// Log battle start to Redis (simplified)
	go func() {
		message := fmt.Sprintf("Savaş başladı: %s vs %s", lightSideName, darkSideName)
		if err := LogBattleStart(ctx, battle, message); err != nil {
			log.Printf("Failed to log battle start: %v", err)
		}
	}()

    if battle.Status == BattleStatusInProgress {
        // Publish battle started event
        go PublishBattleStartedEvent(
            battle.ID,
            battle.BattleType,
            0, // No single warrior ID in team battles
            "Team Battle",
            "",
            "",
            "",
        )
    }

	return battle, participants, nil
}

// GetNextParticipant gets the next alive participant in turn order
func (s *Service) GetNextParticipant(ctx context.Context, battle *Battle) (*BattleParticipant, int, error) {
	// Get all alive participants ordered by creation (turn order)
	filter := bson.M{
		"battle_id": battle.ID,
		"is_alive":  true,
	}

	opts := options.Find().SetSort(bson.M{"created_at": 1}) // Oldest first = turn order
	cursor, err := BattleParticipantColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find participants: %w", err)
	}
	defer cursor.Close(ctx)

	var participants []BattleParticipant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, 0, fmt.Errorf("failed to decode participants: %w", err)
	}

	if len(participants) == 0 {
		return nil, 0, errors.New("no alive participants found")
	}

	// Get participant at current index (with wrap-around)
	index := battle.CurrentParticipantIndex % len(participants)
	participant := participants[index]

	return &participant, index, nil
}

// CheckTeamStatus checks if a team has any alive participants
func (s *Service) CheckTeamStatus(ctx context.Context, battleID primitive.ObjectID, side TeamSide) (bool, error) {
	filter := bson.M{
		"battle_id": battleID,
		"side":      side,
		"is_alive":  true,
	}

	count, err := BattleParticipantColl.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to count team participants: %w", err)
	}

	return count > 0, nil
}

