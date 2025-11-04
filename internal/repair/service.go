package repair

import (
    "context"
    "fmt"
    "time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ComputeRepairCost calculates repair cost based on durability and RBAC role
// RBAC-based pricing: Emperors get 50% discount, Kings get 25% discount
func (s *Service) ComputeRepairCost(ctx context.Context, currentDur, maxDur int, role string) int {
    missing := maxDur - currentDur
    if missing < 0 { missing = 0 }
    baseCost := missing * 2
    
    // Apply RBAC-based discounts
    switch role {
    case "light_emperor", "dark_emperor":
        return baseCost / 2 // 50% discount for emperors
    case "light_king", "dark_king":
        return baseCost * 3 / 4 // 25% discount for kings
    default:
        return baseCost // Full price for regular warriors
    }
}

func (s *Service) CreateRepairOrder(ctx context.Context, ownerType, ownerID, itemID, itemType string, cost int) (*RepairOrder, error) {
    ro := &RepairOrder{
        OwnerType: ownerType,
        OwnerID: ownerID,
        ItemType: itemType,
        Cost: cost,
        Status: RepairStatusPending,
        CreatedAt: time.Now(),
    }
    if itemType == "weapon" {
        ro.WeaponID = itemID
    } else if itemType == "armor" {
        ro.ArmorID = itemID
    }
    if err := s.repo.CreateOrder(ctx, ro); err != nil { return nil, err }
    return ro, nil
}

func (s *Service) CompleteRepair(ctx context.Context, orderID uint) error {
    now := time.Now()
    return s.repo.CompleteOrder(ctx, orderID, now)
}

func (s *Service) ListOrders(ctx context.Context, ownerType, ownerID string) ([]RepairOrder, error) {
    return s.repo.ListOrders(ctx, ownerType, ownerID)
}

var ErrInvalidInput = fmt.Errorf("invalid input")


