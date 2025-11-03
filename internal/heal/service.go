package heal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"network-sec-micro/internal/heal/dto"
)

// Service handles healing business logic with CQRS pattern
type Service struct {
	repo Repository
}

// NewService creates a new heal service
func NewService() *Service {
	return &Service{
		repo: GetRepository(),
	}
}

// GetHealPackageByType returns heal package by type with role validation
func GetHealPackageByType(healType HealType, role string) (HealPackage, error) {
	var packageInfo HealPackage
	switch healType {
	case HealTypeFull:
		packageInfo = FullHealPackage
	case HealTypePartial:
		packageInfo = PartialHealPackage
	case HealTypeEmperorFull:
		packageInfo = EmperorFullHealPackage
	case HealTypeEmperorPartial:
		packageInfo = EmperorPartialHealPackage
	case HealTypeDragon:
		packageInfo = DragonHealPackage
	default:
		return HealPackage{}, errors.New("invalid heal type")
	}

	// Check role permission
	normalizedRole := normalizeRole(role)
	if !canUsePackage(normalizedRole, packageInfo.RequiredRole) {
		return HealPackage{}, fmt.Errorf("role '%s' cannot use %s package (requires %s)", role, healType, packageInfo.RequiredRole)
	}

	return packageInfo, nil
}

// normalizeRole normalizes role string for comparison
func normalizeRole(role string) string {
	if role == "light_emperor" || role == "dark_emperor" {
		return "emperor"
	}
	if role == "dragon" {
		return "dragon"
	}
	return "warrior"
}

// canUsePackage checks if role can use the package
func canUsePackage(userRole, requiredRole string) bool {
	if requiredRole == "warrior" {
		return true // All roles can use warrior packages
	}
	return userRole == requiredRole
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// PurchaseHeal processes a healing purchase (Command) - supports Warrior, Dragon, and Enemy
func (s *Service) PurchaseHeal(ctx context.Context, cmd dto.PurchaseHealCommand) (*HealingRecord, error) {
	healType := HealType(cmd.HealType)
	participantID := cmd.ParticipantID
	participantType := cmd.ParticipantType
	participantRole := cmd.ParticipantRole

	// Validate participant type
	if participantType != "warrior" && participantType != "dragon" && participantType != "enemy" {
		return nil, fmt.Errorf("invalid participant type: %s (must be warrior, dragon, or enemy)", participantType)
	}

	// Get heal package with role validation
	packageInfo, err := GetHealPackageByType(healType, participantRole)
	if err != nil {
		return nil, err
	}

	var participantName string
	var currentHP, maxHP int
	var isHealing bool
	var healingUntil *time.Time

	// Get participant info based on type
	switch participantType {
	case "warrior":
		warriorID, err := strconv.ParseUint(participantID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid warrior ID: %w", err)
		}

		warrior, err := GetWarriorByID(ctx, uint(warriorID))
		if err != nil {
			return nil, fmt.Errorf("failed to get warrior: %w", err)
		}

		participantName = warrior.Username
		currentHP = int(warrior.CurrentHp)
		maxHP = int(warrior.MaxHp)
		if maxHP == 0 {
			maxHP = int(warrior.TotalPower) * 10
			if maxHP < 100 {
				maxHP = 100
			}
		}

		// Check if warrior is already healing
		isHealing, healingUntil, err = CheckWarriorHealingState(ctx, uint(warriorID))
		if err != nil {
			log.Printf("Warning: Could not check healing state: %v", err)
		}

	case "dragon":
		dragon, err := GetDragonByID(ctx, participantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get dragon: %w", err)
		}

		if !dragon.IsAlive {
			return nil, errors.New("dragon is not alive and cannot be healed")
		}

		participantName = dragon.Name
		currentHP = int(dragon.Health)
		maxHP = int(dragon.MaxHealth)

		// Check if dragon is already healing
		isHealing, healingUntil, err = CheckDragonHealingState(ctx, participantID)
		if err != nil {
			log.Printf("Warning: Could not check healing state: %v", err)
		}

	case "enemy":
		enemy, err := GetEnemyByID(ctx, participantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get enemy: %w", err)
		}

		participantName = enemy.Name
		currentHP = int(enemy.Health)
		maxHealth := int(enemy.MaxHealth)
		if maxHealth == 0 {
			maxHealth = currentHP // Fallback to current HP if max not set
		}
		maxHP = maxHealth

		// Check if enemy is already healing
		isHealing, healingUntil, err = CheckEnemyHealingState(ctx, participantID)
		if err != nil {
			log.Printf("Warning: Could not check healing state: %v", err)
		}

	default:
		return nil, fmt.Errorf("unsupported participant type: %s", participantType)
	}

	// Check if already healing
	if isHealing && healingUntil != nil {
		if time.Now().Before(*healingUntil) {
			remaining := time.Until(*healingUntil).Seconds()
			return nil, fmt.Errorf("%s is already healing. Remaining time: %.0f seconds", participantType, remaining)
		}
		// Healing time passed, clear state
		switch participantType {
		case "warrior":
			warriorID, _ := strconv.ParseUint(participantID, 10, 32)
			_ = SetWarriorHealingState(ctx, uint(warriorID), false, nil)
		case "dragon":
			_ = SetDragonHealingState(ctx, participantID, false, nil)
		case "enemy":
			_ = SetEnemyHealingState(ctx, participantID, false, nil)
		}
	}

	// Calculate healing amount based on type
	hpBefore := currentHP
	var healedAmount int
	var hpAfter int

	switch healType {
	case HealTypeFull, HealTypeEmperorFull:
		healedAmount = maxHP - hpBefore
		hpAfter = maxHP
	case HealTypePartial, HealTypeEmperorPartial:
		healedAmount = hpBefore / 2
		hpAfter = hpBefore + healedAmount
		if hpAfter > maxHP {
			hpAfter = maxHP
			healedAmount = maxHP - hpBefore
		}
	case HealTypeDragon:
		healedAmount = maxHP - hpBefore
		hpAfter = maxHP
	default:
		return nil, errors.New("invalid heal type for calculation")
	}

	// Check if healing is needed
	if hpBefore >= maxHP && (healType == HealTypeFull || healType == HealTypeEmperorFull || healType == HealTypeDragon) {
		return nil, fmt.Errorf("%s already at full HP", participantType)
	}
	if healedAmount <= 0 {
		return nil, errors.New("no healing needed")
	}

	// Deduct coins (only for warriors, dragons/enemies are NPCs)
	if err := DeductCoinsForParticipant(ctx, participantID, participantType, int64(packageInfo.Price), fmt.Sprintf("heal_%s", healType)); err != nil {
		return nil, fmt.Errorf("failed to deduct coins: %w", err)
	}

	// Calculate healing completion time
	now := time.Now()
	completedAt := now.Add(time.Duration(packageInfo.Duration) * time.Second)

	// Set healing state
	switch participantType {
	case "warrior":
		warriorID, _ := strconv.ParseUint(participantID, 10, 32)
		if err := SetWarriorHealingState(ctx, uint(warriorID), true, &completedAt); err != nil {
			log.Printf("Warning: Failed to set warrior healing state: %v", err)
		}
	case "dragon":
		if err := SetDragonHealingState(ctx, participantID, true, &completedAt); err != nil {
			log.Printf("Warning: Failed to set dragon healing state: %v", err)
		}
	case "enemy":
		if err := SetEnemyHealingState(ctx, participantID, true, &completedAt); err != nil {
			log.Printf("Warning: Failed to set enemy healing state: %v", err)
		}
	}

	// Convert warriorID for legacy compatibility
	var warriorID uint
	if participantType == "warrior" {
		warriorID, _ = strconv.ParseUint(participantID, 10, 32)
		warriorID = uint(warriorID)
	}

	// Create healing record
	record := &HealingRecord{
		ID:             fmt.Sprintf("%s-%s-%d", participantType, participantID, now.Unix()),
		ParticipantID:  participantID,
		ParticipantType: participantType,
		ParticipantName: participantName,
		WarriorID:      warriorID,
		WarriorName:    participantName, // Legacy field
		HealType:       healType,
		HealedAmount:   healedAmount,
		HPBefore:       hpBefore,
		HPAfter:        hpAfter,
		CoinsSpent:     packageInfo.Price,
		Duration:       packageInfo.Duration,
		CompletedAt:    &completedAt,
		CreatedAt:      now,
	}

	// Save to database
	if err := s.repo.SaveHealingRecord(ctx, record); err != nil {
		log.Printf("Warning: Failed to save healing record: %v", err)
	}

	// Log healing started to Redis
	if err := LogHealingStarted(ctx, record); err != nil {
		log.Printf("Warning: Failed to log healing started: %v", err)
	}

	// Schedule HP update after duration with progress logging
	go func() {
		remaining := packageInfo.Duration
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		done := make(chan bool)
		go func() {
			time.Sleep(time.Duration(packageInfo.Duration) * time.Second)
			done <- true
		}()

		for {
			select {
			case <-ticker.C:
				remaining -= 5
				if remaining < 0 {
					remaining = 0
				}
				progress := float64(packageInfo.Duration-remaining) / float64(packageInfo.Duration) * 100.0
				if err := LogHealingProgress(context.Background(), warriorID, participantName, healType, remaining, packageInfo.Duration, progress); err != nil {
					log.Printf("Warning: Failed to log healing progress: %v", err)
				}
			case <-done:
				// Apply HP update
				var updateErr error
				switch participantType {
				case "warrior":
					warriorIDUint, _ := strconv.ParseUint(participantID, 10, 32)
					updateErr = UpdateWarriorHP(context.Background(), uint(warriorIDUint), int32(hpAfter))
				case "dragon":
					updateErr = UpdateDragonHP(context.Background(), participantID, int32(hpAfter))
				case "enemy":
					updateErr = UpdateEnemyHP(context.Background(), participantID, int32(hpAfter))
				}

				if updateErr != nil {
					log.Printf("Failed to apply healing HP after duration: %v", updateErr)
					_ = LogHealingFailed(context.Background(), warriorID, participantName, healType, fmt.Sprintf("Failed to update HP: %v", updateErr))
				} else {
				// Clear healing state
				switch participantType {
				case "warrior":
					warriorIDUint, _ := strconv.ParseUint(participantID, 10, 32)
					_ = SetWarriorHealingState(context.Background(), uint(warriorIDUint), false, nil)
					case "dragon":
						_ = SetDragonHealingState(context.Background(), participantID, false, nil)
					case "enemy":
						_ = SetEnemyHealingState(context.Background(), participantID, false, nil)
					}

					// Update record completion
					record.CompletedAt = &time.Time{}
					*record.CompletedAt = time.Now()

					// Log healing completed
					if err := LogHealingCompleted(context.Background(), record); err != nil {
						log.Printf("Warning: Failed to log healing completed: %v", err)
					}
					log.Printf("Healing completed for %s %s: HP updated to %d", participantType, participantID, hpAfter)
				}
				return
			}
		}
	}()

	log.Printf("Healing started: %s=%s, type=%s, will heal=%d, hp: %d->%d, coins=%d, duration=%ds",
		participantType, participantID, healType, healedAmount, hpBefore, hpAfter, packageInfo.Price, packageInfo.Duration)

	return record, nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetHealingHistory retrieves healing history for a participant (Query)
func (s *Service) GetHealingHistory(ctx context.Context, query dto.GetHealingHistoryQuery) ([]*HealingRecord, error) {
	// For now, we'll use warriorID for backward compatibility
	// In the future, we should support participant_id and participant_type
	return s.repo.GetHealingHistory(ctx, query.WarriorID)
}
