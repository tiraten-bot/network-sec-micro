package heal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
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
	warriorRole := normalizeRole(role)
	if !canUsePackage(warriorRole, packageInfo.RequiredRole) {
		return HealPackage{}, fmt.Errorf("role '%s' cannot use %s package (requires %s)", role, healType, packageInfo.RequiredRole)
	}

	return packageInfo, nil
}

// normalizeRole normalizes role string for comparison
func normalizeRole(role string) string {
	if role == "light_emperor" || role == "dark_emperor" {
		return "emperor"
	}
	// Assume dragon role is passed as "dragon" or similar
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

// PurchaseHeal processes a healing purchase (Command)
func (s *Service) PurchaseHeal(ctx context.Context, cmd dto.PurchaseHealCommand) (*HealingRecord, error) {
	healType := HealType(cmd.HealType)
	warriorID := cmd.WarriorID
	battleID := cmd.BattleID
	warriorRole := cmd.WarriorRole
	// Get warrior info
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	// Check if warrior is already healing
	isHealing, healingUntil, err := CheckWarriorHealingState(ctx, warriorID)
	if err != nil {
		log.Printf("Warning: Could not check healing state: %v", err)
	}
	if isHealing && healingUntil != nil {
		if time.Now().Before(*healingUntil) {
			remaining := time.Until(*healingUntil).Seconds()
			return nil, fmt.Errorf("warrior is already healing. Remaining time: %.0f seconds", remaining)
		}
		// Healing time passed, clear state
		_ = SetWarriorHealingState(ctx, warriorID, false, nil)
	}

	// Get current HP from battle logs (if battleID provided) or warrior service
	currentHP := int(warrior.CurrentHp)
	if currentHP == 0 && battleID != "" {
		hp, err := GetBattleLogLastHP(ctx, battleID, warriorID)
		if err != nil {
			log.Printf("Warning: Could not get HP from battle logs: %v. Using warrior default", err)
		} else {
			currentHP = hp
		}
	}

	// Calculate max HP from total power
	maxHP := int(warrior.MaxHp)
	if maxHP == 0 {
		maxHP = int(warrior.TotalPower) * 10
		if maxHP < 100 {
			maxHP = 100 // Minimum HP
		}
	}

	// Get heal package with role validation
	packageInfo, err := GetHealPackageByType(healType, warriorRole)
	if err != nil {
		return nil, err
	}

	// Calculate healing amount based on type
	hpBefore := currentHP
	var healedAmount int
	var hpAfter int

	switch healType {
	case HealTypeFull, HealTypeEmperorFull:
		// Full heal: restore to max HP
		healedAmount = maxHP - hpBefore
		hpAfter = maxHP
	case HealTypePartial, HealTypeEmperorPartial:
		// Partial heal: 50% of current HP
		healedAmount = hpBefore / 2
		hpAfter = hpBefore + healedAmount
		if hpAfter > maxHP {
			hpAfter = maxHP
			healedAmount = maxHP - hpBefore
		}
	case HealTypeDragon:
		// Dragon heal: full heal but very slow
		healedAmount = maxHP - hpBefore
		hpAfter = maxHP
	default:
		return nil, errors.New("invalid heal type for calculation")
	}

	// Check if healing is needed
	if hpBefore >= maxHP && (healType == HealTypeFull || healType == HealTypeEmperorFull || healType == HealTypeDragon) {
		return nil, errors.New("warrior already at full HP")
	}
	if healedAmount <= 0 {
		return nil, errors.New("no healing needed")
	}

	// Deduct coins
	if err := DeductCoins(ctx, warriorID, int64(packageInfo.Price), fmt.Sprintf("heal_%s", healType)); err != nil {
		return nil, fmt.Errorf("failed to deduct coins: %w", err)
	}

	// Calculate healing completion time
	now := time.Now()
	completedAt := now.Add(time.Duration(packageInfo.Duration) * time.Second)

	// Set warrior healing state (HP will be updated after duration)
	// For now, we'll schedule the HP update
	// In production, you'd use a background job or timer
	if err := SetWarriorHealingState(ctx, warriorID, true, &completedAt); err != nil {
		log.Printf("Warning: Failed to set warrior healing state: %v", err)
	}

	// Create healing record (HP will be applied after duration)
	record := &HealingRecord{
		ID:           fmt.Sprintf("%d-%d", warriorID, now.Unix()),
		WarriorID:    warriorID,
		WarriorName:  warrior.Username,
		HealType:     healType,
		HealedAmount: healedAmount,
		HPBefore:     hpBefore,
		HPAfter:      hpAfter,
		CoinsSpent:   packageInfo.Price,
		Duration:     packageInfo.Duration,
		CompletedAt:  &completedAt,
		CreatedAt:    now,
	}

	// Save to database
	if err := GetRepository().SaveHealingRecord(ctx, record); err != nil {
		log.Printf("Warning: Failed to save healing record: %v", err)
	}

	// Schedule HP update after duration (in production, use background job)
	go func() {
		time.Sleep(time.Duration(packageInfo.Duration) * time.Second)
		if err := UpdateWarriorHP(context.Background(), warriorID, int32(hpAfter)); err != nil {
			log.Printf("Failed to apply healing HP after duration: %v", err)
		} else {
			// Clear healing state
			_ = SetWarriorHealingState(context.Background(), warriorID, false, nil)
			log.Printf("Healing completed for warrior %d: HP updated to %d", warriorID, hpAfter)
		}
	}()

	log.Printf("Healing started: warrior=%d, type=%s, will heal=%d, hp: %d->%d, coins=%d, duration=%ds",
		warriorID, healType, healedAmount, hpBefore, hpAfter, packageInfo.Price, packageInfo.Duration)

	return record, nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetHealingHistory retrieves healing history for a warrior (Query)
func (s *Service) GetHealingHistory(ctx context.Context, query dto.GetHealingHistoryQuery) ([]*HealingRecord, error) {
	return s.repo.GetHealingHistory(ctx, query.WarriorID)
}

