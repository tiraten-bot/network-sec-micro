package repair

import (
    "context"
    "fmt"
    "time"
)

type Service struct{}

func NewService() *Service { return &Service{} }

// ComputeRepairCost simple rule: cost = (max_durability - durability) * 2
func (s *Service) ComputeRepairCost(ctx context.Context, currentDur, maxDur int) int {
    missing := maxDur - currentDur
    if missing < 0 { missing = 0 }
    return missing * 2
}

func (s *Service) CreateRepairOrder(ctx context.Context, ownerType, ownerID, weaponID string, cost int) (*RepairOrder, error) {
    ro := &RepairOrder{OwnerType: ownerType, OwnerID: ownerID, WeaponID: weaponID, Cost: cost, Status: RepairStatusPending, CreatedAt: time.Now()}
    if err := GetDB().WithContext(ctx).Create(ro).Error; err != nil { return nil, err }
    return ro, nil
}

func (s *Service) CompleteRepair(ctx context.Context, orderID uint) error {
    var ro RepairOrder
    if err := GetDB().WithContext(ctx).First(&ro, orderID).Error; err != nil { return err }
    now := time.Now()
    ro.Status = RepairStatusCompleted
    ro.CompletedAt = &now
    return GetDB().WithContext(ctx).Save(&ro).Error
}

func (s *Service) ListOrders(ctx context.Context, ownerType, ownerID string) ([]RepairOrder, error) {
    var out []RepairOrder
    if err := GetDB().WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Order("id DESC").Find(&out).Error; err != nil {
        return nil, err
    }
    return out, nil
}

var ErrInvalidInput = fmt.Errorf("invalid input")


