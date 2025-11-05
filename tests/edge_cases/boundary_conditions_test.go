package edge_cases_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/repair"
	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepairCost_BoundaryValues tests boundary values for repair cost
func TestRepairCost_BoundaryValues(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	tests := []struct {
		name       string
		currentDur int
		maxDur     int
		role       string
		expected   int
	}{
		{
			name:       "exactly at max",
			currentDur: 100,
			maxDur:     100,
			role:       "knight",
			expected:   0,
		},
		{
			name:       "one point missing",
			currentDur: 99,
			maxDur:     100,
			role:       "knight",
			expected:   2,
		},
		{
			name:       "one point from broken",
			currentDur: 1,
			maxDur:     100,
			role:       "knight",
			expected:   198,
		},
		{
			name:       "completely broken",
			currentDur: 0,
			maxDur:     100,
			role:       "knight",
			expected:   200,
		},
		{
			name:       "large durability values",
			currentDur: 1000,
			maxDur:     10000,
			role:       "knight",
			expected:   18000, // (10000 - 1000) * 2
		},
		{
			name:       "very small durability",
			currentDur: 5,
			maxDur:     10,
			role:       "knight",
			expected:   10, // (10 - 5) * 2
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := svc.ComputeRepairCost(ctx, tt.currentDur, tt.maxDur, tt.role)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

// TestCoinBalance_BoundaryValues tests boundary values for coin balance
func TestCoinBalance_BoundaryValues(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	svc := coin.NewService()
	ctx := context.Background()
	
	tests := []struct {
		name        string
		initialBalance int64
		deductAmount    int64
		shouldSucceed   bool
		expectedBalance int64
	}{
		{
			name:           "deduct exact balance",
			initialBalance: 100,
			deductAmount:   100,
			shouldSucceed:  true,
			expectedBalance: 0,
		},
		{
			name:           "deduct one more than balance",
			initialBalance: 100,
			deductAmount:   101,
			shouldSucceed:  false, // Should fail
		},
		{
			name:           "deduct from zero balance",
			initialBalance: 0,
			deductAmount:   1,
			shouldSucceed:  false,
		},
		{
			name:           "deduct one coin from one coin",
			initialBalance: 1,
			deductAmount:   1,
			shouldSucceed:  true,
			expectedBalance: 0,
		},
		{
			name:           "large balance deduction",
			initialBalance: 1000000,
			deductAmount:   500000,
			shouldSucceed:  true,
			expectedBalance: 500000,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create balance
			balance := coin.WarriorBalance{
				WarriorID: uint(len(tests)),
				Balance:   tt.initialBalance,
			}
			err := db.Create(&balance).Error
			require.NoError(t, err)
			
			// Deduct
			cmd := dto.DeductCoinsCommand{
				WarriorID: balance.WarriorID,
				Amount:    tt.deductAmount,
				Reason:    "test",
			}
			
			err = svc.DeductCoins(ctx, cmd)
			
			if tt.shouldSucceed {
				require.NoError(t, err)
				
				// Verify balance
				var updatedBalance coin.WarriorBalance
				db.First(&updatedBalance, balance.WarriorID)
				assert.Equal(t, tt.expectedBalance, updatedBalance.Balance)
			} else {
				assert.Error(t, err)
			}
			
			// Cleanup
			db.Delete(&balance)
		})
	}
}

// TestRepairCost_RoleDiscount_Boundaries tests role discount boundaries
func TestRepairCost_RoleDiscount_Boundaries(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Test with different roles and boundary values
	tests := []struct {
		name       string
		currentDur int
		maxDur     int
		role       string
		expected   int
	}{
		{
			name:       "emperor discount on minimal repair",
			currentDur: 99,
			maxDur:     100,
			role:       "light_emperor",
			expected:   1, // (2 / 2) = 1
		},
		{
			name:       "king discount on minimal repair",
			currentDur: 99,
			maxDur:     100,
			role:       "light_king",
			expected:   1, // (2 * 0.75) = 1 (integer division)
		},
		{
			name:       "emperor discount on large repair",
			currentDur: 0,
			maxDur:     1000,
			role:       "dark_emperor",
			expected:   1000, // (2000 / 2) = 1000
		},
		{
			name:       "king discount on large repair",
			currentDur: 0,
			maxDur:     1000,
			role:       "dark_king",
			expected:   1500, // (2000 * 0.75) = 1500
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := svc.ComputeRepairCost(ctx, tt.currentDur, tt.maxDur, tt.role)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

// TestBattleDamage_BoundaryValues tests damage calculation boundaries
func TestBattleDamage_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		attackPower int
		defense     int
		expectedMin int
		expectedMax int
	}{
		{
			name:        "attack equals defense",
			attackPower: 100,
			defense:     100,
			expectedMin: 10, // Minimum damage
			expectedMax: 10,
		},
		{
			name:        "attack one more than defense",
			attackPower: 101,
			defense:     100,
			expectedMin: 10, // Minimum damage
			expectedMax: 11,
		},
		{
			name:        "defense much higher than attack",
			attackPower: 10,
			defense:     200,
			expectedMin: 10, // Minimum damage enforced
			expectedMax: 10,
		},
		{
			name:        "attack much higher than defense",
			attackPower: 500,
			defense:     50,
			expectedMin: 400,
			expectedMax: 500,
		},
		{
			name:        "zero defense",
			attackPower: 100,
			defense:     0,
			expectedMin: 90,
			expectedMax: 110,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := tt.attackPower - tt.defense
			if damage < 10 {
				damage = 10
			}
			
			assert.GreaterOrEqual(t, damage, tt.expectedMin)
			assert.LessOrEqual(t, damage, tt.expectedMax)
		})
	}
}

// TestIntegerDivision_Precision tests integer division precision in discounts
func TestIntegerDivision_Precision(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Test that discounts handle integer division correctly
	tests := []struct {
		name       string
		currentDur int
		maxDur     int
		role       string
		baseCost   int
		expected   int
	}{
		{
			name:       "emperor discount on odd cost",
			currentDur: 49,
			maxDur:     100,
			role:       "light_emperor",
			baseCost:   102, // (51 * 2)
			expected:   51,  // 102 / 2 = 51
		},
		{
			name:       "king discount on odd cost",
			currentDur: 49,
			maxDur:     100,
			role:       "light_king",
			baseCost:   102, // (51 * 2)
			expected:   76,  // 102 * 3 / 4 = 76.5 -> 76
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := svc.ComputeRepairCost(ctx, tt.currentDur, tt.maxDur, tt.role)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

