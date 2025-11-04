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
// StartBattle is now in service_team.go
// Attack is now in service_attack.go

// OLD FUNCTIONS BELOW - TO BE REMOVED AFTER MIGRATION
func (s *Service) oldStartBattle(cmd dto.StartBattleCommand) (*Battle, error) {
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
		OpponentMaxHP: cmd.OpponentMaxHP,
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

    battle.ID = result.InsertedID.(primitive.ObjectID).Hex()

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

	// Log battle start to Redis
	go func() {
		if err := LogBattleStart(ctx, battle, fmt.Sprintf("Battle started: %s vs %s", battle.WarriorName, battle.OpponentName)); err != nil {
			log.Printf("Failed to log battle start: %v", err)
		}
	}()

	// Publish battle started event
    go PublishBattleStartedEvent(
        battle.ID,
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

	// Handle team battle vs legacy single battle
	if battle.BattleType == BattleTypeTeam {
		return s.performTeamBattleAttack(ctx, &battle, cmd)
	}

	// Legacy single battle logic below
	// Get warrior info
	warrior, err := GetWarriorByUsername(ctx, cmd.WarriorName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get warrior info: %w", err)
	}

	// Get warrior's weapons for bonus damage
	weaponBonus := 0
	var usedWeaponID string
	if ws, err := ListWeaponsByOwner(ctx, "warrior", cmd.WarriorName); err == nil {
		maxD := 0
		for _, w := range ws {
			if w.IsBroken { continue }
			if int(w.Damage) > maxD { 
				maxD = int(w.Damage)
				usedWeaponID = w.Id 
			}
		}
		weaponBonus = maxD
		if usedWeaponID != "" { 
			_, _ = ApplyWeaponWear(ctx, usedWeaponID, 1) 
		}
	}

	// Get opponent's armors for defense bonus (if opponent is warrior/enemy/dragon)
	opponentDefenseBonus := 0
	var usedArmorID string
	// For legacy single battles, check opponent type
	if battle.OpponentType == "warrior" || battle.OpponentType == "enemy" || battle.OpponentType == "dragon" {
		if armors, err := ListArmorsByOwner(ctx, battle.OpponentType, battle.OpponentID); err == nil {
			maxDef := 0
			for _, a := range armors {
				if a.IsBroken { continue }
				if int(a.Defense) > maxDef { 
					maxDef = int(a.Defense)
					usedArmorID = a.Id 
				}
			}
			opponentDefenseBonus = maxDef
			if usedArmorID != "" { 
				_, _ = ApplyArmorWear(ctx, usedArmorID, 1) 
			}
		}
	}

	// Warrior attacks opponent
	warriorPower := int(warrior.TotalPower)
	targetDefense := battle.OpponentDefense() + opponentDefenseBonus
	damage := s.calculateDamage(warriorPower + weaponBonus, targetDefense)
	
	// Critical hit chance (10%)
	isCritical := rand.Float64() < 0.1
	if isCritical {
		damage = int(float64(damage) * 1.5)
	}

	targetHPBefore := battle.OpponentHP
	battle.OpponentHP -= damage
	if battle.OpponentHP < 0 {
		battle.OpponentHP = 0
	}

	// Increment turn
	battle.CurrentTurn++
	battle.UpdatedAt = time.Now()

	// Create turn record for warrior attack
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

	// Log warrior attack to Redis
	eventType := "warrior_attack"
	message := fmt.Sprintf("%s attacks %s for %d damage", battle.WarriorName, battle.OpponentName, damage)
	if isCritical {
		eventType = "critical_hit"
		message = fmt.Sprintf("⚔️ CRITICAL HIT! %s attacks %s for %d damage!", battle.WarriorName, battle.OpponentName, damage)
	}
	
	// Create a temp battle object with HP before attack for logging
	tempBattle := battle
	tempBattle.OpponentHP = targetHPBefore
	if err := LogBattleTurn(ctx, battle.ID, turn, &tempBattle, eventType, message); err != nil {
		log.Printf("Warning: failed to log battle turn to Redis: %v", err)
	}

	// Check if opponent is defeated
	if battle.OpponentHP <= 0 {
		return s.completeBattle(ctx, &battle, BattleResultVictory, battle.WarriorName, battle.WarriorID)
	}

	// Opponent counter-attacks asynchronously (non-blocking)
	// This simulates real-time combat where opponent responds
	go func(battleCopy Battle) {
		// Small delay to simulate opponent thinking/reacting (500ms - 1s)
		delay := time.Duration(500+rand.Intn(500)) * time.Millisecond
		time.Sleep(delay)

		oppCtx := context.Background()
		
		// Re-fetch battle to get latest state
		var currentBattle Battle
		err := BattleColl.FindOne(oppCtx, bson.M{"_id": battleCopy.ID}).Decode(&currentBattle)
		if err != nil {
			log.Printf("Failed to fetch battle for opponent attack: %v", err)
			return
		}

		// Check if battle is still in progress
		if currentBattle.Status != BattleStatusInProgress {
			return
		}

		// Get warrior's armor for defense bonus
		warriorDefenseBonus := 0
		var warriorArmorID string
		if armors, err := ListArmorsByOwner(ctx, "warrior", currentBattle.WarriorName); err == nil {
			maxDef := 0
			for _, a := range armors {
				if a.IsBroken { continue }
				if int(a.Defense) > maxDef { 
					maxDef = int(a.Defense)
					warriorArmorID = a.Id 
				}
			}
			warriorDefenseBonus = maxDef
			if warriorArmorID != "" { 
				_, _ = ApplyArmorWear(ctx, warriorArmorID, 1) 
			}
		}

		// Opponent attacks
		opponentDamage := s.calculateOpponentDamage(&currentBattle)
		// Apply warrior's armor defense bonus
		if warriorDefenseBonus > 0 {
			opponentDamage = opponentDamage - warriorDefenseBonus
			if opponentDamage < 1 {
				opponentDamage = 1 // Minimum 1 damage
			}
		}
		opponentCritical := rand.Float64() < 0.05 // 5% crit for opponent
		if opponentCritical {
			opponentDamage = int(float64(opponentDamage) * 1.5)
		}

		warriorHPBefore := currentBattle.WarriorHP
		currentBattle.WarriorHP -= opponentDamage
		if currentBattle.WarriorHP < 0 {
			currentBattle.WarriorHP = 0
		}

		// Record opponent turn
		opponentTurn := &BattleTurn{
			BattleID:      currentBattle.ID,
			TurnNumber:    currentBattle.CurrentTurn,
			AttackerID:    currentBattle.OpponentID,
			AttackerName:  currentBattle.OpponentName,
			AttackerType:  currentBattle.OpponentType,
			TargetID:      fmt.Sprintf("%d", currentBattle.WarriorID),
			TargetName:    currentBattle.WarriorName,
			TargetType:    "warrior",
			DamageDealt:   opponentDamage,
			CriticalHit:   opponentCritical,
			TargetHPAfter: currentBattle.WarriorHP,
			CreatedAt:     time.Now(),
		}

		_, err = BattleTurnColl.InsertOne(oppCtx, opponentTurn)
		if err != nil {
			log.Printf("Failed to record opponent turn: %v", err)
			return
		}

		// Log opponent attack to Redis
		oppEventType := "opponent_attack"
		oppMessage := fmt.Sprintf("%s counter-attacks %s for %d damage", currentBattle.OpponentName, currentBattle.WarriorName, opponentDamage)
		if opponentCritical {
			oppEventType = "critical_hit"
			oppMessage = fmt.Sprintf("💥 CRITICAL HIT! %s counter-attacks %s for %d damage!", currentBattle.OpponentName, currentBattle.WarriorName, opponentDamage)
		}

		tempBattleForLog := currentBattle
		tempBattleForLog.WarriorHP = warriorHPBefore
		if err := LogBattleTurn(oppCtx, currentBattle.ID, opponentTurn, &tempBattleForLog, oppEventType, oppMessage); err != nil {
			log.Printf("Warning: failed to log opponent turn to Redis: %v", err)
		}

		// Check if warrior is defeated
		if currentBattle.WarriorHP <= 0 {
			s.completeBattle(oppCtx, &currentBattle, BattleResultDefeat, currentBattle.OpponentName, 0)
			return
		}

		// Update battle
		updateData := bson.M{
			"warrior_hp":  currentBattle.WarriorHP,
			"opponent_hp": currentBattle.OpponentHP,
			"updated_at":  time.Now(),
		}

		_, err = BattleColl.UpdateOne(oppCtx, bson.M{"_id": currentBattle.ID}, bson.M{"$set": updateData})
		if err != nil {
			log.Printf("Failed to update battle after opponent attack: %v", err)
		}
	}(battle)

	// Update battle (warrior attack only, opponent will update async)
	updateData := bson.M{
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

		// Add coins to warrior via gRPC
		ctx := context.Background()
		go func() {
			if err := AddCoins(ctx, battle.WarriorID, int64(battle.CoinsEarned), fmt.Sprintf("battle_victory_%s", battle.ID.Hex())); err != nil {
				log.Printf("Failed to add coins to warrior %d after battle victory: %v", battle.WarriorID, err)
			}
		}()
	} else if result == BattleResultDefeat {
		// Deduct coins from warrior if lost (penalty)
		penalty := 25 // Base penalty
		ctx := context.Background()
		go func() {
			if err := DeductCoins(ctx, battle.WarriorID, int64(penalty), fmt.Sprintf("battle_defeat_penalty_%s", battle.ID.Hex())); err != nil {
				// Log but don't fail - penalty might fail if insufficient balance
				log.Printf("Failed to deduct penalty coins from warrior %d after battle defeat: %v", battle.WarriorID, err)
			}
		}()
		battle.CoinsEarned = -penalty // Negative to indicate loss
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

	// Log battle end to Redis
	go func() {
		endMessage := fmt.Sprintf("Battle completed. Result: %s. Winner: %s", result, battle.WinnerName)
		if err := LogBattleEnd(ctx, battle, endMessage); err != nil {
			log.Printf("Failed to log battle end: %v", err)
		}
	}()

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

// performTeamBattleAttack handles team battle participant-based attacks
func (s *Service) performTeamBattleAttack(ctx context.Context, battle *Battle, cmd dto.AttackCommand) (*Battle, *BattleTurn, error) {
	if cmd.AttackerID == "" || cmd.TargetID == "" {
		return nil, nil, errors.New("attacker_id and target_id are required for team battles")
	}

	// Get attacker and target participants
	attacker, err := GetRepository().GetParticipantByIDs(ctx, battle.ID, cmd.AttackerID)
	if err != nil {
		return nil, nil, fmt.Errorf("attacker participant not found: %w", err)
	}

	target, err := GetRepository().GetParticipantByIDs(ctx, battle.ID, cmd.TargetID)
	if err != nil {
		return nil, nil, fmt.Errorf("target participant not found: %w", err)
	}

	// Validate they're on different sides
	if attacker.Side == target.Side {
		return nil, nil, errors.New("attacker and target must be on different sides")
	}

	// Validate attacker is alive
	if !attacker.IsAlive {
		return nil, nil, errors.New("attacker is not alive")
	}

	// Validate target is alive
	if !target.IsAlive {
		return nil, nil, errors.New("target is not alive")
	}

	// Get attacker's weapons for bonus damage
	weaponBonus := 0
	var usedWeaponID string
	if string(attacker.Type) == "warrior" || string(attacker.Type) == "enemy" || string(attacker.Type) == "dragon" {
		if ws, err := ListWeaponsByOwner(ctx, string(attacker.Type), attacker.ParticipantID); err == nil {
			maxD := 0
			for _, w := range ws {
				if w.IsBroken { continue }
				if int(w.Damage) > maxD { 
					maxD = int(w.Damage)
					usedWeaponID = w.Id 
				}
			}
			weaponBonus = maxD
			if usedWeaponID != "" { 
				_, _ = ApplyWeaponWear(ctx, usedWeaponID, 1) 
			}
		}
	}

	// Get target's armors for defense bonus
	targetDefenseBonus := 0
	var usedArmorID string
	if string(target.Type) == "warrior" || string(target.Type) == "enemy" || string(target.Type) == "dragon" {
		if armors, err := ListArmorsByOwner(ctx, string(target.Type), target.ParticipantID); err == nil {
			maxDef := 0
			for _, a := range armors {
				if a.IsBroken { continue }
				if int(a.Defense) > maxDef { 
					maxDef = int(a.Defense)
					usedArmorID = a.Id 
				}
			}
			targetDefenseBonus = maxDef
			if usedArmorID != "" { 
				_, _ = ApplyArmorWear(ctx, usedArmorID, 1) 
			}
		}
	}

	// Calculate damage
	attackerPower := attacker.AttackPower + weaponBonus
	targetDefense := target.Defense + targetDefenseBonus
	damage := s.calculateDamage(attackerPower, targetDefense)

	// Critical hit chance (10%)
	isCritical := rand.Float64() < 0.1
	if isCritical {
		damage = int(float64(damage) * 1.5)
	}

	// Apply damage
	targetHPBefore := target.HP
	target.HP -= damage
	if target.HP < 0 {
		target.HP = 0
	}

	// Check if target is defeated
	targetDefeated := target.HP <= 0
	if targetDefeated {
		target.IsAlive = false
		target.IsDefeated = true
		now := time.Now()
		target.DefeatedAt = &now
	}

	// Update target participant
	updateTarget := map[string]interface{}{
		"hp": target.HP,
		"is_alive": target.IsAlive,
		"is_defeated": target.IsDefeated,
		"updated_at": time.Now(),
	}
	if target.DefeatedAt != nil {
		updateTarget["defeated_at"] = target.DefeatedAt
	}
	if err := GetRepository().UpdateParticipantByIDs(ctx, battle.ID, target.ParticipantID, updateTarget); err != nil {
		return nil, nil, fmt.Errorf("failed to update target participant: %w", err)
	}

	// Increment battle turn
	battle.CurrentTurn++
	battle.CurrentParticipantIndex++
	battle.UpdatedAt = time.Now()

	// Create turn record
	turn := &BattleTurn{
		BattleID:      battle.ID,
		TurnNumber:    battle.CurrentTurn,
		AttackerID:    attacker.ParticipantID,
		AttackerName:  attacker.Name,
		AttackerType:  attacker.Type,
		AttackerSide:  attacker.Side,
		TargetID:      target.ParticipantID,
		TargetName:    target.Name,
		TargetType:    target.Type,
		TargetSide:    target.Side,
		DamageDealt:   damage,
		CriticalHit:   isCritical,
		TargetHPBefore: targetHPBefore,
		TargetHPAfter: target.HP,
		TargetDefeated: targetDefeated,
		CreatedAt:     time.Now(),
	}

	if err := GetRepository().InsertTurn(ctx, turn); err != nil {
		return nil, nil, fmt.Errorf("failed to record turn: %w", err)
	}

	// Check if battle is complete (one side has no alive participants)
	lightAlive, err := GetRepository().CountAliveBySide(ctx, battle.ID, TeamSideLight)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count light side: %w", err)
	}
	darkAlive, err := GetRepository().CountAliveBySide(ctx, battle.ID, TeamSideDark)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count dark side: %w", err)
	}

	if lightAlive == 0 {
		// Dark side wins
		return s.completeTeamBattle(ctx, battle, BattleResultDarkVictory)
	} else if darkAlive == 0 {
		// Light side wins
		return s.completeTeamBattle(ctx, battle, BattleResultLightVictory)
	}

	// Update battle
	updateData := map[string]interface{}{
		"current_turn": battle.CurrentTurn,
		"current_participant_index": battle.CurrentParticipantIndex,
		"updated_at": battle.UpdatedAt,
	}
	if err := GetRepository().UpdateBattleFields(ctx, battle.ID, updateData); err != nil {
		return nil, nil, fmt.Errorf("failed to update battle: %w", err)
	}

	return battle, turn, nil
}

// completeTeamBattle marks a team battle as completed
func (s *Service) completeTeamBattle(ctx context.Context, battle *Battle, result BattleResult) (*Battle, *BattleTurn, error) {
	now := time.Now()
	battle.Status = BattleStatusCompleted
	battle.Result = result
	battle.CompletedAt = &now

	if result == BattleResultLightVictory {
		battle.WinnerSide = TeamSideLight
	} else if result == BattleResultDarkVictory {
		battle.WinnerSide = TeamSideDark
	}

	updateData := map[string]interface{}{
		"status": battle.Status,
		"result": battle.Result,
		"winner_side": battle.WinnerSide,
		"completed_at": battle.CompletedAt,
		"updated_at": now,
	}
	if err := GetRepository().UpdateBattleFields(ctx, battle.ID, updateData); err != nil {
		return nil, nil, fmt.Errorf("failed to complete battle: %w", err)
	}

	// Publish battle completed event (simplified signature for team battles)
	go func() {
		_ = PublishBattleCompletedEvent(
			battle.ID,
			string(battle.BattleType),
			0, // No single warrior ID in team battles
			"Team Battle",
			string(result),
			string(battle.WinnerSide),
			0, // Coins earned (calculated separately)
			0, // Experience gained (calculated separately)
			battle.CurrentTurn,
		)
	}()

	return battle, nil, nil
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

// GetBattlesByWarrior gets battles for a warrior (or all if warriorID is 0)
func (s *Service) GetBattlesByWarrior(query dto.GetBattlesByWarriorQuery) ([]Battle, int64, error) {
	ctx := context.Background()

	filter := bson.M{}
	if query.WarriorID > 0 {
		filter["warrior_id"] = query.WarriorID
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

