package battle

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/battle/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetBattle gets a battle by ID
func (s *Service) GetBattle(query dto.GetBattleQuery) (*Battle, []*BattleParticipant, []*BattleParticipant, error) {
	ctx := context.Background()

	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": query.BattleID}).Decode(&battle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, nil, errors.New("battle not found")
		}
		return nil, nil, nil, fmt.Errorf("failed to get battle: %w", err)
	}

	// Get participants
	lightParticipants, err := s.GetBattleParticipants(ctx, query.BattleID, "light")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get light participants: %w", err)
	}

	darkParticipants, err := s.GetBattleParticipants(ctx, query.BattleID, "dark")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get dark participants: %w", err)
	}

	return &battle, lightParticipants, darkParticipants, nil
}

// GetBattlesByWarrior gets battles for a warrior (or all if warriorID is 0)
func (s *Service) GetBattlesByWarrior(query dto.GetBattlesByWarriorQuery) ([]Battle, int64, error) {
	ctx := context.Background()

	filter := bson.M{}
	
	// If warriorID specified, find battles where warrior participated
	if query.WarriorID > 0 {
		// Find participant with this warrior ID
		participantCursor, err := BattleParticipantColl.Find(ctx, bson.M{
			"participant_id": fmt.Sprintf("%d", query.WarriorID),
			"type": ParticipantTypeWarrior,
		}, options.Find().SetProjection(bson.M{"battle_id": 1}))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to find warrior battles: %w", err)
		}
		defer participantCursor.Close(ctx)

		var battleIDs []primitive.ObjectID
		for participantCursor.Next(ctx) {
			var p BattleParticipant
			if err := participantCursor.Decode(&p); err == nil {
				battleIDs = append(battleIDs, p.BattleID)
			}
		}

		if len(battleIDs) > 0 {
			filter["_id"] = bson.M{"$in": battleIDs}
		} else {
			// No battles found
			return []Battle{}, 0, nil
		}
	}

	if query.Status != "all" && query.Status != "" {
		filter["status"] = query.Status
	}

	// Count total
	total, err := BattleColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count battles: %w", err)
	}

	// Apply pagination
	opts := options.Find()
	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}
	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}
	opts.SetSort(bson.M{"created_at": -1}) // Latest first

	cursor, err := BattleColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find battles: %w", err)
	}
	defer cursor.Close(ctx)

	var battles []Battle
	if err := cursor.All(ctx, &battles); err != nil {
		return nil, 0, fmt.Errorf("failed to decode battles: %w", err)
	}

	return battles, total, nil
}

// GetBattleTurns gets turns for a battle
func (s *Service) GetBattleTurns(query dto.GetBattleTurnsQuery) ([]BattleTurn, error) {
	ctx := context.Background()

	filter := bson.M{"battle_id": query.BattleID}

	opts := options.Find()
	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}
	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}
	opts.SetSort(bson.M{"turn_number": 1}) // Ascending order

	cursor, err := BattleTurnColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find turns: %w", err)
	}
	defer cursor.Close(ctx)

	var turns []BattleTurn
	if err := cursor.All(ctx, &turns); err != nil {
		return nil, fmt.Errorf("failed to decode turns: %w", err)
	}

	return turns, nil
}

// GetBattleStats gets battle statistics for a warrior
func (s *Service) GetBattleStats(query dto.GetBattleStatsQuery) (*dto.BattleStatsResponse, error) {
	ctx := context.Background()

	// Find battles where warrior participated
	participantCursor, err := BattleParticipantColl.Find(ctx, bson.M{
		"participant_id": fmt.Sprintf("%d", query.WarriorID),
		"type": ParticipantTypeWarrior,
	}, options.Find().SetProjection(bson.M{"battle_id": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to find warrior battles: %w", err)
	}
	defer participantCursor.Close(ctx)

	var battleIDs []primitive.ObjectID
	for participantCursor.Next(ctx) {
		var p BattleParticipant
		if err := participantCursor.Decode(&p); err == nil {
			battleIDs = append(battleIDs, p.BattleID)
		}
	}

	if len(battleIDs) == 0 {
		// No battles
		return &dto.BattleStatsResponse{
			WarriorID: query.WarriorID,
		}, nil
	}

	filter := bson.M{
		"_id":    bson.M{"$in": battleIDs},
		"status": "completed",
	}

	if query.BattleType != "all" && query.BattleType != "" {
		filter["battle_type"] = query.BattleType
	}

	cursor, err := BattleColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find battles: %w", err)
	}
	defer cursor.Close(ctx)

	stats := &dto.BattleStatsResponse{
		WarriorID: query.WarriorID,
	}

	var battles []Battle
	if err := cursor.All(ctx, &battles); err != nil {
		return nil, fmt.Errorf("failed to decode battles: %w", err)
	}

	for _, battle := range battles {
		stats.TotalBattles++

		// Determine if warrior's team won
		warriorSide := TeamSideLight // Default, will check from participant
		
		// Get warrior's side from participants
		var warriorParticipant BattleParticipant
		err := BattleParticipantColl.FindOne(ctx, bson.M{
			"battle_id":      battle.ID,
			"participant_id": fmt.Sprintf("%d", query.WarriorID),
		}).Decode(&warriorParticipant)
		if err == nil {
			warriorSide = warriorParticipant.Side
		}

		switch battle.Result {
		case BattleResultLightVictory:
			if warriorSide == TeamSideLight {
				stats.Wins++
			} else {
				stats.Losses++
			}
		case BattleResultDarkVictory:
			if warriorSide == TeamSideDark {
				stats.Wins++
			} else {
				stats.Losses++
			}
		case BattleResultDraw:
			stats.Draws++
		}

		if battle.BattleType == BattleTypeTeam {
			stats.TeamBattles++
		}

		// Add coins and experience (for this warrior)
		if coins, ok := battle.CoinsEarned[fmt.Sprintf("%d", query.WarriorID)]; ok {
			stats.TotalCoinsEarned += coins
		}
		if exp, ok := battle.ExperienceGained[fmt.Sprintf("%d", query.WarriorID)]; ok {
			stats.TotalExperience += exp
		}
	}

	// Calculate win rate
	if stats.TotalBattles > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.TotalBattles) * 100
	}

	return stats, nil
}

