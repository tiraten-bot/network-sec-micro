package repair_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"network-sec-micro/internal/repair"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository is a simple in-memory repository for testing
type mockRepository struct {
	orders []repair.RepairOrder
	nextID uint
}

func (m *mockRepository) CreateOrder(ctx context.Context, order *repair.RepairOrder) error {
	m.nextID++
	order.ID = m.nextID
	m.orders = append(m.orders, *order)
	return nil
}

func (m *mockRepository) CompleteOrder(ctx context.Context, orderID uint, completedAt time.Time) error {
	for i := range m.orders {
		if m.orders[i].ID == orderID {
			m.orders[i].Status = repair.RepairStatusCompleted
			m.orders[i].CompletedAt = &completedAt
			return nil
		}
	}
	return fmt.Errorf("order not found")
}

func (m *mockRepository) ListOrders(ctx context.Context, ownerType, ownerID string) ([]repair.RepairOrder, error) {
	var result []repair.RepairOrder
	for _, order := range m.orders {
		if order.OwnerType == ownerType && order.OwnerID == ownerID {
			result = append(result, order)
		}
	}
	return result, nil
}

func TestComputeRepairCost_RegularWarrior(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Full repair: 100 durability missing
	cost := svc.ComputeRepairCost(ctx, 0, 100, "knight")
	assert.Equal(t, 200, cost) // 100 * 2
	
	// Partial repair: 50 durability missing
	cost = svc.ComputeRepairCost(ctx, 50, 100, "knight")
	assert.Equal(t, 100, cost) // 50 * 2
	
	// No repair needed
	cost = svc.ComputeRepairCost(ctx, 100, 100, "knight")
	assert.Equal(t, 0, cost)
}

func TestComputeRepairCost_EmperorDiscount(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Light emperor gets 50% discount
	cost := svc.ComputeRepairCost(ctx, 0, 100, "light_emperor")
	assert.Equal(t, 100, cost) // 200 * 0.5
	
	// Dark emperor gets 50% discount
	cost = svc.ComputeRepairCost(ctx, 0, 100, "dark_emperor")
	assert.Equal(t, 100, cost) // 200 * 0.5
}

func TestComputeRepairCost_KingDiscount(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Light king gets 25% discount
	cost := svc.ComputeRepairCost(ctx, 0, 100, "light_king")
	assert.Equal(t, 150, cost) // 200 * 0.75
	
	// Dark king gets 25% discount
	cost = svc.ComputeRepairCost(ctx, 0, 100, "dark_king")
	assert.Equal(t, 150, cost) // 200 * 0.75
}

func TestComputeRepairCost_NegativeDurability(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Current durability higher than max (shouldn't happen, but test edge case)
	cost := svc.ComputeRepairCost(ctx, 150, 100, "knight")
	assert.Equal(t, 0, cost) // missing is negative, clamped to 0
}

func TestCreateRepairOrder_Weapon(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 200)
	
	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "warrior", order.OwnerType)
	assert.Equal(t, "warrior1", order.OwnerID)
	assert.Equal(t, "weapon1", order.WeaponID)
	assert.Equal(t, "weapon", order.ItemType)
	assert.Equal(t, 200, order.Cost)
	assert.Equal(t, repair.RepairStatusPending, order.Status)
	assert.NotZero(t, order.ID)
}

func TestCreateRepairOrder_Armor(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "armor1", "armor", 150)
	
	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, "armor1", order.ArmorID)
	assert.Equal(t, "armor", order.ItemType)
}

func TestCompleteRepair_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Create order
	order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 200)
	require.NoError(t, err)
	
	// Complete order
	err = svc.CompleteRepair(ctx, order.ID)
	assert.NoError(t, err)
	
	// Verify order is completed
	orders, err := svc.ListOrders(ctx, "warrior", "warrior1")
	require.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, repair.RepairStatusCompleted, orders[0].Status)
	assert.NotNil(t, orders[0].CompletedAt)
}

func TestListOrders_Success(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Create multiple orders for same owner
	_, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 200)
	require.NoError(t, err)
	_, err = svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon2", "weapon", 150)
	require.NoError(t, err)
	_, err = svc.CreateRepairOrder(ctx, "warrior", "warrior2", "weapon3", "weapon", 100)
	require.NoError(t, err)
	
	// List orders for warrior1
	orders, err := svc.ListOrders(ctx, "warrior", "warrior1")
	
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	for _, order := range orders {
		assert.Equal(t, "warrior1", order.OwnerID)
	}
}

func TestListOrders_Empty(t *testing.T) {
	repo := &mockRepository{}
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	orders, err := svc.ListOrders(ctx, "warrior", "nonexistent")
	
	require.NoError(t, err)
	assert.Empty(t, orders)
}

func TestRepairOrder_StatusConstants(t *testing.T) {
	assert.Equal(t, repair.RepairOrderStatus("pending"), repair.RepairStatusPending)
	assert.Equal(t, repair.RepairOrderStatus("completed"), repair.RepairStatusCompleted)
	assert.Equal(t, repair.RepairOrderStatus("failed"), repair.RepairStatusFailed)
}

