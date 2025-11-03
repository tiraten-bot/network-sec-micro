package battle

import (
    "context"
    "errors"
    "fmt"
    "log"
    "math/rand"
    "time"

    "network-sec-micro/internal/battle/dto"
)

// Attack performs an attack in a team-based battle
func (s *Service) Attack(cmd dto.AttackCommand) (*Battle, *BattleTurn, error) {
	ctx := context.Background()

	// Get battle
    if cmd.BattleID == "" { return nil, nil, errors.New("invalid battle ID") }
    battle, err := GetRepository().GetBattleByID(ctx, cmd.BattleID)
    if err != nil { return nil, nil, fmt.Errorf("failed to get battle: %w", err) }

	// Validate battle status
	if battle.Status != BattleStatusInProgress {
		return nil, nil, errors.New("battle is not in progress")
	}

	// Validate turn limit
	if battle.CurrentTurn >= battle.MaxTurns {
		return s.handleTeamBattleTimeout(ctx, &battle)
	}

	// Get attacker participant
    attacker, err := GetRepository().GetParticipantByIDs(ctx, battle.ID, cmd.AttackerID)
    if err != nil { return nil, nil, fmt.Errorf("failed to get attacker: %w", err) }

	// Get target participant
    target, err := GetRepository().GetParticipantByIDs(ctx, battle.ID, cmd.TargetID)
    if err != nil { return nil, nil, fmt.Errorf("failed to get target: %w", err) }

	// Validate: attacker and target must be on different sides
	if attacker.Side == target.Side {
		return nil, nil, errors.New("cannot attack teammate")
	}

	// Calculate damage
	damage := s.calculateParticipantDamage(attacker.AttackPower, target.Defense)

	// Critical hit chance (10% for warriors, 5% for others)
	critChance := 0.1
	if attacker.Type != ParticipantTypeWarrior {
		critChance = 0.05
	}
	isCritical := rand.Float64() < critChance
	if isCritical {
		damage = int(float64(damage) * 1.5)
	}

    targetHPBefore := target.HP
    target.HP -= damage
    if target.HP < 0 { target.HP = 0 }

	// Track damage for kill attribution
	killTracker.AddDamage(battleID, target.ParticipantID, attacker.ParticipantID)

	// Check if target is defeated
	targetDefeated := target.HP <= 0
	if targetDefeated {
		target.IsAlive = false
		target.IsDefeated = true
		now := time.Now()
		target.DefeatedAt = &now

		// Get all participants who damaged this target before it died
		killerIDs := killTracker.GetKillers(battleID, target.ParticipantID)
		
		// Create target copy for async logging
		targetCopy := target
		battleIDCopy := battleID
		
		// Log simplified death event to Redis (async)
		go func() {
			killerParticipants, _ := GetParticipantObjects(context.Background(), battleIDCopy, killerIDs)
			if err := LogParticipantDefeated(context.Background(), battleIDCopy, &targetCopy, killerParticipants); err != nil {
				log.Printf("Warning: failed to log participant death: %v", err)
			}
		}()

		// If target is a dragon, check for revival possibility
		if target.Type == ParticipantTypeDragon {
			go func() {
				if err := s.HandleDragonDeathInBattle(ctx, battleID, &targetCopy); err != nil {
					log.Printf("Warning: failed to handle dragon death: %v", err)
				}
			}()
		}

		// If attacker is a dragon and killed a warrior, trigger Wraith of Dragon spell via battlespell service
		if attacker.Type == ParticipantTypeDragon && target.Type == ParticipantTypeWarrior && targetDefeated {
			go func() {
				// Call battlespell service via gRPC
				battlespellClient := GetBattlespellClient()
				if battlespellClient != nil {
					req := &pbBattleSpell.TriggerWraithOfDragonRequest{
						BattleId: battleID.Hex(),
					}
					resp, err := battlespellClient.TriggerWraithOfDragon(context.Background(), req)
					if err != nil {
						log.Printf("Warning: failed to trigger wraith of dragon via battlespell: %v", err)
					} else if resp.Triggered {
						log.Printf("Wraith of Dragon triggered: %s destroyed (count: %d/25)", resp.DestroyedWarriorId, resp.WraithCount)
					}
				}
			}()
		}
	}

	// Update target participant
	updateTarget := bson.M{
		"hp":         target.HP,
		"is_alive":   target.IsAlive,
		"is_defeated": target.IsDefeated,
		"updated_at": time.Now(),
	}
	if target.DefeatedAt != nil {
		updateTarget["defeated_at"] = target.DefeatedAt
	}

    if err := GetRepository().UpdateParticipantByIDs(ctx, battle.ID, target.ParticipantID, updateTarget); err != nil {
        return nil, nil, fmt.Errorf("failed to update target: %w", err)
    }

	// Increment turn
	battle.CurrentTurn++
	battle.CurrentParticipantIndex++
	battle.UpdatedAt = time.Now()

	// Create turn record
    turn := &BattleTurn{
        BattleID:        battle.ID,
		TurnNumber:      battle.CurrentTurn,
		AttackerID:      attacker.ParticipantID,
		AttackerName:    attacker.Name,
		AttackerType:    attacker.Type,
		AttackerSide:    attacker.Side,
		TargetID:        target.ParticipantID,
		TargetName:      target.Name,
		TargetType:      target.Type,
		TargetSide:       target.Side,
		DamageDealt:     damage,
		CriticalHit:     isCritical,
		TargetHPBefore:  targetHPBefore,
		TargetHPAfter:   target.HP,
		TargetDefeated:  targetDefeated,
		CreatedAt:       time.Now(),
	}
    if err := GetRepository().InsertTurn(ctx, turn); err != nil { return nil, nil, fmt.Errorf("failed to record turn: %w", err) }

	// Note: We only log to Redis when a participant is defeated (see above)
	// No need to log every attack, only deaths are logged in simplified format

	// Check if a team has been eliminated
    lightAliveCount, err := GetRepository().CountAliveBySide(ctx, battle.ID, TeamSideLight)
	if err != nil {
		log.Printf("Warning: failed to check light team status: %v", err)
	}
    darkAliveCount, err := GetRepository().CountAliveBySide(ctx, battle.ID, TeamSideDark)
	if err != nil {
		log.Printf("Warning: failed to check dark team status: %v", err)
	}
    lightAlive := lightAliveCount > 0
    darkAlive := darkAliveCount > 0

	// Determine battle result
	if !lightAlive && !darkAlive {
		// Both teams eliminated - draw
		return s.completeTeamBattle(ctx, &battle, BattleResultDraw, TeamSideLight, allParticipants)
	} else if !lightAlive {
		// Dark side wins
		return s.completeTeamBattle(ctx, &battle, BattleResultDarkVictory, TeamSideDark, allParticipants)
	} else if !darkAlive {
		// Light side wins
		return s.completeTeamBattle(ctx, &battle, BattleResultLightVictory, TeamSideLight, allParticipants)
	}

	// Update battle
    updateData := map[string]interface{}{
        "current_turn": battle.CurrentTurn,
        "current_participant_index": battle.CurrentParticipantIndex,
        "updated_at": time.Now(),
    }
    if err := GetRepository().UpdateBattleFields(ctx, battle.ID, updateData); err != nil { return nil, nil, fmt.Errorf("failed to update battle: %w", err) }

	return &battle, turn, nil
}

// calculateParticipantDamage calculates damage dealt by attacker to target
func (s *Service) calculateParticipantDamage(attackerPower, targetDefense int) int {
	baseDamage := attackerPower - targetDefense
	if baseDamage < 10 {
		baseDamage = 10 // Minimum damage
	}

	// Add randomness (±20%)
	randomFactor := 0.8 + (rand.Float64() * 0.4)
	return int(float64(baseDamage) * randomFactor)
}

// completeTeamBattle marks a team battle as completed
func (s *Service) completeTeamBattle(ctx context.Context, battle *Battle, result BattleResult, winnerSide TeamSide, participants []*BattleParticipant) (*Battle, *BattleTurn, error) {
	now := time.Now()
	battle.Status = BattleStatusCompleted
	battle.Result = result
	battle.WinnerSide = winnerSide
	battle.CompletedAt = &now

	// Initialize reward maps
	battle.CoinsEarned = make(map[string]int)
	battle.ExperienceGained = make(map[string]int)

	// Calculate rewards for winning team
	if result == BattleResultLightVictory || result == BattleResultDarkVictory {
		baseCoins := 50 + (battle.CurrentTurn * 2)
		baseExp := 100 + (battle.CurrentTurn * 5)

		for _, p := range participants {
			if p.Side == winnerSide && p.IsAlive {
				// Survivors get rewards
				battle.CoinsEarned[p.ParticipantID] = baseCoins
				battle.ExperienceGained[p.ParticipantID] = baseExp

				// Distribute coins via gRPC (only for warriors)
				if p.Type == ParticipantTypeWarrior {
					warriorID := p.ParticipantID // Assuming it's a numeric string
					go func(pid string, coins int64) {
						// Parse warrior ID if needed and call AddCoins
						// For now, we'll log it
						log.Printf("Warrior %s earned %d coins from team battle victory", pid, coins)
					}(warriorID, int64(baseCoins))
				}
			}
		}
	} else {
		// Draw - smaller rewards for all survivors
		for _, p := range participants {
			if p.IsAlive {
				battle.CoinsEarned[p.ParticipantID] = 25
				battle.ExperienceGained[p.ParticipantID] = 50
			}
		}
	}

	updateData := bson.M{
		"status":         battle.Status,
		"result":         battle.Result,
		"winner_side":    battle.WinnerSide,
		"completed_at":   battle.CompletedAt,
		"coins_earned":   battle.CoinsEarned,
		"experience_gained": battle.ExperienceGained,
		"updated_at":     time.Now(),
	}

	_, err := BattleColl.UpdateOne(ctx, bson.M{"_id": battle.ID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to complete battle: %w", err)
	}

	// Log battle end to Redis (simplified)
	go func() {
		var endMessage string
		if result == BattleResultDraw {
			endMessage = fmt.Sprintf("Savaş berabere bitti")
		} else {
			endMessage = fmt.Sprintf("Savaş bitti. Kazanan: %s", winnerSide)
		}
		if err := LogBattleEnd(ctx, battle, endMessage); err != nil {
			log.Printf("Failed to log battle end: %v", err)
		}
		// Clear kill tracker for this battle
		killTracker.ClearKills(battle.ID)
	}()

	// Publish battle completed event
	go PublishBattleCompletedEvent(
		battle.ID,
		string(battle.BattleType),
		0,
		"Team Battle",
		string(result),
		string(winnerSide),
		0, // Total coins (calculated per participant)
		0, // Total exp
		battle.CurrentTurn,
	)

	// Publish wager resolution if configured and approved
	if battle.WagerAmount > 0 && battle.LightEmperorID != "" && battle.DarkEmperorID != "" && battle.LightEmperorApproved && battle.DarkEmperorApproved {
		_ = PublishBattleWagerResolved(battle.ID, winnerSide, battle.WagerAmount, battle.LightEmperorID, battle.DarkEmperorID)
	}

	return battle, nil, nil
}

// handleTeamBattleTimeout handles battle timeout (max turns reached)
func (s *Service) handleTeamBattleTimeout(ctx context.Context, battle *Battle) (*Battle, *BattleTurn, error) {
	// Count alive participants on each side
	lightAlive, err := s.CheckTeamStatus(ctx, battle.ID, TeamSideLight)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check light team: %w", err)
	}

	darkAlive, err := s.CheckTeamStatus(ctx, battle.ID, TeamSideDark)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check dark team: %w", err)
	}

	// Get all participants to calculate HP totals
	participants, _ := s.GetBattleParticipants(ctx, battle.ID, "all")

	var lightTotalHP, darkTotalHP int
	for _, p := range participants {
		if p.IsAlive {
			if p.Side == TeamSideLight {
				lightTotalHP += p.HP
			} else {
				darkTotalHP += p.HP
			}
		}
	}

	var result BattleResult
	var winnerSide TeamSide

	if lightTotalHP > darkTotalHP {
		result = BattleResultLightVictory
		winnerSide = TeamSideLight
	} else if darkTotalHP > lightTotalHP {
		result = BattleResultDarkVictory
		winnerSide = TeamSideDark
	} else {
		result = BattleResultDraw
		winnerSide = TeamSideLight // Doesn't matter for draw
	}

	return s.completeTeamBattle(ctx, battle, result, winnerSide, participants)
}

// GetBattleParticipants gets all participants for a battle
func (s *Service) GetBattleParticipants(ctx context.Context, battleID primitive.ObjectID, sideFilter string) ([]*BattleParticipant, error) {
	filter := bson.M{"battle_id": battleID}
	if sideFilter != "all" && sideFilter != "" {
		filter["side"] = TeamSide(sideFilter)
	}

	cursor, err := BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find participants: %w", err)
	}
	defer cursor.Close(ctx)

	var participants []BattleParticipant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, fmt.Errorf("failed to decode participants: %w", err)
	}

	result := make([]*BattleParticipant, len(participants))
	for i := range participants {
		result[i] = &participants[i]
	}

	return result, nil
}

