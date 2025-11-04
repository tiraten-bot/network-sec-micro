package repair

import (
    "context"
    "fmt"
    "time"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ComputeRepairCost simple rule: cost = (max_durability - durability) * 2
func (s *Service) ComputeRepairCost(ctx context.Context, currentDur, maxDur int) int {
    missing := maxDur - currentDur
    if missing < 0 { missing = 0 }
    return missing * 2
}

func (s *Service) CreateRepairOrder(ctx context.Context, ownerType, ownerID, weaponID string, cost int) (*RepairOrder, error) {
    ro := &RepairOrder{OwnerType: ownerType, OwnerID: ownerID, WeaponID: weaponID, Cost: cost, Status: RepairStatusPending, CreatedAt: time.Now()}
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


