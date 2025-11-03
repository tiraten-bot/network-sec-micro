package heal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
)

// Service handles healing business logic
type Service struct{}

// NewService creates a new heal service
func NewService() *Service {
	return &Service{}
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

// PurchaseHeal processes a healing purchase
func (s *Service) PurchaseHeal(ctx context.Context, warriorID uint, healType HealType, battleID string, warriorRole string) (*HealingRecord, error) {
	// Get warrior info
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	// Check if warrior is already healing
	if warrior.CurrentHp > 0 { // If we have HP info, check healing state
		// TODO: Check IsHealing field from warrior service
		// For now, we'll check healing records for active healing
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

	// Calculate healing amount
	hpBefore := currentHP
	var healedAmount int
	var hpAfter int

	if healType == HealTypeFull {
		// Full heal: restore to max HP
		healedAmount = maxHP - hpBefore
		hpAfter = maxHP
	} else {
		// Partial heal: 50% of current HP
		healedAmount = hpBefore / 2
		hpAfter = hpBefore + healedAmount
		if hpAfter > maxHP {
			hpAfter = maxHP
			healedAmount = maxHP - hpBefore
		}
	}

	// Check if healing is needed
	if hpBefore >= maxHP && healType == HealTypeFull {
		return nil, errors.New("warrior already at full HP")
	}
	if healedAmount <= 0 {
		return nil, errors.New("no healing needed")
	}

	// Deduct coins
	if err := DeductCoins(ctx, warriorID, int64(packageInfo.Price), fmt.Sprintf("heal_%s", healType)); err != nil {
		return nil, fmt.Errorf("failed to deduct coins: %w", err)
	}

	// Update warrior HP via gRPC
	if err := UpdateWarriorHP(ctx, warriorID, hpAfter); err != nil {
		log.Printf("Warning: Failed to update warrior HP via gRPC: %v", err)
		// Continue anyway - healing record will be saved
	}

	// Create healing record
	record := &HealingRecord{
		ID:           fmt.Sprintf("%d-%d", warriorID, time.Now().Unix()),
		WarriorID:    warriorID,
		WarriorName:  warrior.Username,
		HealType:     healType,
		HealedAmount: healedAmount,
		HPBefore:     hpBefore,
		HPAfter:      hpAfter,
		CoinsSpent:   packageInfo.Price,
		CreatedAt:    time.Now(),
	}

	// Save to database
	if err := GetRepository().SaveHealingRecord(ctx, record); err != nil {
		log.Printf("Warning: Failed to save healing record: %v", err)
	}

	log.Printf("Healing applied: warrior=%d, type=%s, healed=%d, hp: %d->%d, coins=%d",
		warriorID, healType, healedAmount, hpBefore, hpAfter, packageInfo.Price)

	return record, nil
}

// GetHealingHistory retrieves healing history for a warrior
func (s *Service) GetHealingHistory(ctx context.Context, warriorID uint) ([]*HealingRecord, error) {
	return GetRepository().GetHealingHistory(ctx, warriorID)
}

