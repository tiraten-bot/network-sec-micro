package repair

import (
    "context"
    "time"
)

// Repository defines CQRS-style access for repair orders
type Repository interface {
    // Commands
    CreateOrder(ctx context.Context, order *RepairOrder) error
    CompleteOrder(ctx context.Context, orderID uint, completedAt time.Time) error

    // Queries
    ListOrders(ctx context.Context, ownerType, ownerID string) ([]RepairOrder, error)
}

type pgRepository struct{}

func (pgRepository) CreateOrder(ctx context.Context, order *RepairOrder) error {
    return GetDB().WithContext(ctx).Create(order).Error
}

func (pgRepository) CompleteOrder(ctx context.Context, orderID uint, completedAt time.Time) error {
    return GetDB().WithContext(ctx).Model(&RepairOrder{}).Where("id = ?", orderID).Updates(map[string]interface{}{
        "status":       RepairStatusCompleted,
        "completed_at": completedAt,
    }).Error
}

func (pgRepository) ListOrders(ctx context.Context, ownerType, ownerID string) ([]RepairOrder, error) {
    var out []RepairOrder
    err := GetDB().WithContext(ctx).
        Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
        Order("id DESC").
        Find(&out).Error
    return out, err
}

// GetRepository provides the default repository implementation
func GetRepository() Repository { return pgRepository{} }


