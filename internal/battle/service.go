package battle

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"network-sec-micro/internal/battle/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Service handles battle business logic with CQRS pattern
type Service struct{}

// NewService creates a new battle service
func NewService() *Service {
	return &Service{}
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// StartBattle creates and starts a new battle
func (s *Service) StartBattle(cmd dto.StartBattleCommand) (*Battle, error) {
	ctx := context.Background()

	// Validate battle type
	battleType := BattleType(cmd.BattleType)
	if battleType != BattleTypeEnemy && battleType != BattleTypeDragon {
		return nil, errors.New("invalid battle type")
	}

	// Get warrior info via gRPC
	warrior, err := GetWarriorByUsername(ctx, cmd.WarriorName)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior info: %w", err)
	}

	// Get opponent info (enemy or dragon) - we'll call their services
	// For now, we assume opponent info is passed in the command
	// In production, we'd make HTTP/gRPC calls to enemy/dragon services

	// Calculate warrior HP (based on total power)
	warriorMaxHP := int(warrior.TotalPower) * 10 // Simple formula
	if warriorMaxHP < 100 {
		warriorMaxHP = 100 // Minimum HP
	}

	// Set max turns (default 20)
	maxTurns := cmd.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 20
	}

	// Create battle
	now := time.Now()
	battle := &Battle{
		BattleType:    battleType,
		WarriorID:     uint(warrior.Id),
		WarriorName:   warrior.Username,
		OpponentID:    cmd.OpponentID,
		OpponentName:  cmd.OpponentName,
		OpponentType:  cmd.OpponentType,
		WarriorHP:     warriorMaxHP,
		WarriorMaxHP:  warriorMaxHP,
		OpponentHP:    cmd.OpponentHP, // Should come from enemy/dragon service
		OpponentMaxHP: cmd.OpponentHP,
		CurrentTurn:   0,
		MaxTurns:      maxTurns,
		Status:        BattleStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := BattleColl.InsertOne(ctx, battle)
	if err != nil {
		return nil, fmt.Errorf("failed to create battle: %w", err)
	}

	battle.ID = result.InsertedID.(primitive.ObjectID)

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
		return nil, fmt.Errorf("failed to start battle: %w", err)
	}

	// Publish battle started event
	go PublishBattleStartedEvent(
		battle.ID.Hex(),
		battle.BattleType,
		battle.WarriorID,
		battle.WarriorName,
		battle.OpponentID,
		battle.OpponentName,
		battle.OpponentType,
	)

	return battle, nil
}

// Attack performs an attack in an active battle
func (s *Service) Attack(cmd dto.AttackCommand) (*Battle, *BattleTurn, error) {
	ctx := context.Background()

	// Get battle
	battleID, err := primitive.ObjectIDFromHex(cmd.BattleID)
	if err != nil {
		return nil, nil, errors.New("invalid battle ID")
	}

	var battle Battle
	err = BattleColl.FindOne(ctx, bson.M{"_id": battleID}).Decode(&battle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, errors.New("battle not found")
		}
		return nil, nil, fmt.Errorf("failed to get battle: %w", err)
	}

	// Validate battle status
	if battle.Status != BattleStatusInProgress {
		return nil, nil, errors.New("battle is not in progress")
	}

	// Validate turn limit
	if battle.CurrentTurn >= battle.MaxTurns {
		// Battle timeout - check winner based on HP
		return s.handleBattleTimeout(ctx, &battle)
	}

	// Get warrior info
	warrior, err := GetWarriorByUsername(ctx, cmd.WarriorName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get warrior info: %w", err)
	}

	// Warrior attacks opponent
	warriorPower := int(warrior.TotalPower)
	damage := s.calculateDamage(warriorPower, battle.OpponentDefense())
	
	// Critical hit chance (10%)
	isCritical := rand.Float64() < 0.1
	if isCritical {
		damage = int(float64(damage) * 1.5)
	}

	battle.OpponentHP -= damage
	if battle.OpponentHP < 0 {
		battle.OpponentHP = 0
	}

	// Increment turn
	battle.CurrentTurn++
	battle.UpdatedAt = time.Now()

	// Create turn record
	turn := &BattleTurn{
		BattleID:     battle.ID,
		TurnNumber:   battle.CurrentTurn,
		AttackerID:   fmt.Sprintf("%d", battle.WarriorID),
		AttackerName: battle.WarriorName,
		AttackerType: "warrior",
		TargetID:     battle.OpponentID,
		TargetName:   battle.OpponentName,
		TargetType:   battle.OpponentType,
		DamageDealt:  damage,
		CriticalHit:  isCritical,
		TargetHPAfter: battle.OpponentHP,
		CreatedAt:    time.Now(),
	}

	_, err = BattleTurnColl.InsertOne(ctx, turn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to record turn: %w", err)
	}

	// Check if opponent is defeated
	if battle.OpponentHP <= 0 {
		return s.completeBattle(ctx, &battle, BattleResultVictory, battle.WarriorName, battle.WarriorID)
	}

	// Opponent counter-attacks (if not defeated)
	opponentDamage := s.calculateOpponentDamage(&battle)
	battle.WarriorHP -= opponentDamage
	if battle.WarriorHP < 0 {
		battle.WarriorHP = 0
	}

	// Record opponent turn
	opponentTurn := &BattleTurn{
		BattleID:      battle.ID,
		TurnNumber:    battle.CurrentTurn,
		AttackerID:    battle.OpponentID,
		AttackerName:  battle.OpponentName,
		AttackerType:  battle.OpponentType,
		TargetID:      fmt.Sprintf("%d", battle.WarriorID),
		TargetName:    battle.WarriorName,
		TargetType:    "warrior",
		DamageDealt:   opponentDamage,
		CriticalHit:   rand.Float64() < 0.05, // 5% crit for opponent
		TargetHPAfter: battle.WarriorHP,
		CreatedAt:     time.Now(),
	}

	_, err = BattleTurnColl.InsertOne(ctx, opponentTurn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to record opponent turn: %w", err)
	}

	// Check if warrior is defeated
	if battle.WarriorHP <= 0 {
		return s.completeBattle(ctx, &battle, BattleResultDefeat, battle.OpponentName, 0)
	}

	// Update battle
	updateData := bson.M{
		"warrior_hp":  battle.WarriorHP,
		"opponent_hp": battle.OpponentHP,
		"current_turn": battle.CurrentTurn,
		"updated_at":   time.Now(),
	}

	_, err = BattleColl.UpdateOne(ctx, bson.M{"_id": battle.ID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update battle: %w", err)
	}

	return &battle, turn, nil
}

// completeBattle marks battle as completed
func (s *Service) completeBattle(ctx context.Context, battle *Battle, result BattleResult, winnerName string, winnerID uint) (*Battle, *BattleTurn, error) {
	now := time.Now()
	battle.Status = BattleStatusCompleted
	battle.Result = result
	battle.CompletedAt = &now
	battle.WinnerName = winnerName
	if winnerID > 0 {
		battle.WinnerID = fmt.Sprintf("%d", winnerID)
	}

	// Calculate rewards if warrior won
	if result == BattleResultVictory {
		// Base rewards
		battle.CoinsEarned = 50 + (battle.CurrentTurn * 5)
		battle.ExperienceGained = 100 + (int(battle.OpponentMaxHP) / 10)
	}

	updateData := bson.M{
		"status":         battle.Status,
		"result":         battle.Result,
		"completed_at":   battle.CompletedAt,
		"winner_name":    battle.WinnerName,
		"winner_id":      battle.WinnerID,
		"coins_earned":   battle.CoinsEarned,
		"experience_gained": battle.ExperienceGained,
		"updated_at":      time.Now(),
	}

	_, err := BattleColl.UpdateOne(ctx, bson.M{"_id": battle.ID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to complete battle: %w", err)
	}

	// Publish battle completed event
	go PublishBattleCompletedEvent(
		battle.ID.Hex(),
		battle.BattleType,
		battle.WarriorID,
		battle.WarriorName,
		string(result),
		battle.WinnerName,
		battle.CoinsEarned,
		battle.ExperienceGained,
		battle.CurrentTurn,
	)

	return battle, nil, nil
}

// handleBattleTimeout handles battle timeout (max turns reached)
func (s *Service) handleBattleTimeout(ctx context.Context, battle *Battle) (*Battle, *BattleTurn, error) {
	// Winner is the one with more HP
	var result BattleResult
	var winnerName string
	var winnerID uint

	if battle.WarriorHP > battle.OpponentHP {
		result = BattleResultVictory
		winnerName = battle.WarriorName
		winnerID = battle.WarriorID
	} else if battle.OpponentHP > battle.WarriorHP {
		result = BattleResultDefeat
		winnerName = battle.OpponentName
	} else {
		result = BattleResultDraw
	}

	return s.completeBattle(ctx, battle, result, winnerName, winnerID)
}

// Helper functions
func (s *Service) calculateDamage(attackerPower, targetDefense int) int {
	baseDamage := attackerPower - targetDefense
	if baseDamage < 10 {
		baseDamage = 10 // Minimum damage
	}

	// Add randomness (±20%)
	randomFactor := 0.8 + (rand.Float64() * 0.4)
	return int(float64(baseDamage) * randomFactor)
}

func (s *Service) calculateOpponentDamage(battle *Battle) int {
	// Simple opponent damage calculation
	// In production, this would fetch opponent stats from enemy/dragon service
	opponentAttack := 50 // Default
	if battle.BattleType == BattleTypeDragon {
		opponentAttack = 100
	}

	// Warrior defense (simplified)
	warriorDefense := 30
	damage := opponentAttack - warriorDefense
	if damage < 10 {
		damage = 10
	}

	randomFactor := 0.8 + (rand.Float64() * 0.4)
	return int(float64(damage) * randomFactor)
}

// OpponentDefense returns opponent defense (simplified)
func (b *Battle) OpponentDefense() int {
	if b.BattleType == BattleTypeDragon {
		return 100
	}
	return 50
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetBattle gets a battle by ID
func (s *Service) GetBattle(query dto.GetBattleQuery) (*Battle, error) {
	ctx := context.Background()

	var battle Battle
	err := BattleColl.FindOne(ctx, bson.M{"_id": query.BattleID}).Decode(&battle)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("battle not found")
		}
		return nil, fmt.Errorf("failed to get battle: %w", err)
	}

	return &battle, nil
}

// GetBattlesByWarrior gets battles for a warrior
func (s *Service) GetBattlesByWarrior(query dto.GetBattlesByWarriorQuery) ([]Battle, int64, error) {
	ctx := context.Background()

	filter := bson.M{"warrior_id": query.WarriorID}
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

	filter := bson.M{"warrior_id": query.WarriorID, "status": "completed"}
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

		switch battle.Result {
		case BattleResultVictory:
			stats.Wins++
		case BattleResultDefeat:
			stats.Losses++
		case BattleResultDraw:
			stats.Draws++
		}

		switch battle.BattleType {
		case BattleTypeEnemy:
			stats.EnemyBattles++
		case BattleTypeDragon:
			stats.DragonBattles++
		}

		stats.TotalCoinsEarned += battle.CoinsEarned
		stats.TotalExperience += battle.ExperienceGained
	}

	// Calculate win rate
	if stats.TotalBattles > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.TotalBattles) * 100
	}

	return stats, nil
}

