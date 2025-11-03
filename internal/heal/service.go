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

// PurchaseHeal processes a healing purchase
func (s *Service) PurchaseHeal(ctx context.Context, warriorID uint, healType HealType, battleID string) (*HealingRecord, error) {
	// Get warrior info
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	// Get current HP from battle logs (if battleID provided)
	currentHP := 0
	if battleID != "" {
		hp, err := GetBattleLogLastHP(ctx, battleID, warriorID)
		if err != nil {
			log.Printf("Warning: Could not get HP from battle logs: %v. Using warrior default", err)
			// Fallback: assume warrior has some HP (could be calculated from total_power)
			// For now, we'll need to get HP from warrior service or assume it's 0
			currentHP = 0
		} else {
			currentHP = hp
		}
	}

	// Calculate max HP from total power (if not available from warrior service)
	maxHP := int(warrior.TotalPower) * 10
	if maxHP < 100 {
		maxHP = 100 // Minimum HP
	}

	// Select heal package
	var packageInfo HealPackage
	switch healType {
	case HealTypeFull:
		packageInfo = FullHealPackage
	case HealTypePartial:
		packageInfo = PartialHealPackage
	default:
		return nil, errors.New("invalid heal type")
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
	// TODO: Implement database storage for healing records
	// For now, return empty slice
	return []*HealingRecord{}, nil
}

