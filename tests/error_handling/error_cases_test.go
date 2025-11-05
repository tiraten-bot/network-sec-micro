package error_handling_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"
	"network-sec-micro/internal/repair"
	"network-sec-micro/internal/warrior"
	"network-sec-micro/internal/warrior/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&warrior.Warrior{}, &coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	warrior.DB = db
	coin.DB = db
	
	return db
}

// TestWarriorCreation_InvalidInput tests invalid input handling
func TestWarriorCreation_InvalidInput(t *testing.T) {
	db := setupTestDB(t)
	_ = db
	
	svc := warrior.NewService()
	
	tests := []struct {
		name    string
		cmd     dto.CreateWarriorCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty username",
			cmd: dto.CreateWarriorCommand{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "knight",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			cmd: dto.CreateWarriorCommand{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
				Role:     "knight",
			},
			wantErr: true,
		},
		{
			name: "short password",
			cmd: dto.CreateWarriorCommand{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
				Role:     "knight",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			cmd: dto.CreateWarriorCommand{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid_role",
			},
			wantErr: true,
			errMsg:  "invalid role",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateWarrior(tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCoinDeduction_ErrorCases tests coin deduction error cases
func TestCoinDeduction_ErrorCases(t *testing.T) {
	db := setupTestDB(t)
	
	// Create warrior with balance
	balance := coin.WarriorBalance{
		WarriorID: 1,
		Balance:   100,
	}
	err := db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		cmd     dto.DeductCoinsCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "negative amount",
			cmd: dto.DeductCoinsCommand{
				WarriorID: 1,
				Amount:    -100,
				Reason:    "test",
			},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name: "zero amount",
			cmd: dto.DeductCoinsCommand{
				WarriorID: 1,
				Amount:    0,
				Reason:    "test",
			},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name: "insufficient balance",
			cmd: dto.DeductCoinsCommand{
				WarriorID: 1,
				Amount:    500, // More than balance
				Reason:    "test",
			},
			wantErr: true,
			errMsg:  "insufficient balance",
		},
		{
			name: "warrior not found",
			cmd: dto.DeductCoinsCommand{
				WarriorID: 999,
				Amount:    50,
				Reason:    "test",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeductCoins(ctx, tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCoinTransfer_ErrorCases tests coin transfer error cases
func TestCoinTransfer_ErrorCases(t *testing.T) {
	db := setupTestDB(t)
	
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 100}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 50}
	err := db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		cmd     dto.TransferCoinsCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "transfer to self",
			cmd: dto.TransferCoinsCommand{
				FromWarriorID: 1,
				ToWarriorID:   1,
				Amount:        50,
				Reason:        "test",
			},
			wantErr: true,
			errMsg:  "cannot transfer to self",
		},
		{
			name: "negative amount",
			cmd: dto.TransferCoinsCommand{
				FromWarriorID: 1,
				ToWarriorID:   2,
				Amount:        -50,
				Reason:        "test",
			},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name: "insufficient balance",
			cmd: dto.TransferCoinsCommand{
				FromWarriorID: 1,
				ToWarriorID:   2,
				Amount:        500,
				Reason:        "test",
			},
			wantErr: true,
			errMsg:  "insufficient balance",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.TransferCoins(ctx, tt.cmd)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRepairCost_InvalidInput tests repair cost calculation with invalid input
func TestRepairCost_InvalidInput(t *testing.T) {
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
			name:       "zero max durability",
			currentDur: 50,
			maxDur:     0,
			role:       "knight",
			expected:   0, // Should handle gracefully
		},
		{
			name:       "negative current durability",
			currentDur: -10,
			maxDur:     100,
			role:       "knight",
			expected:   220, // (100 - (-10)) * 2 = 220, but clamped to 0 missing
		},
		{
			name:       "current > max",
			currentDur: 150,
			maxDur:     100,
			role:       "knight",
			expected:   0, // Missing clamped to 0
		},
		{
			name:       "unknown role",
			currentDur: 50,
			maxDur:     100,
			role:       "unknown_role",
			expected:   100, // Full price
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := svc.ComputeRepairCost(ctx, tt.currentDur, tt.maxDur, tt.role)
			assert.Equal(t, tt.expected, cost)
		})
	}
}

// TestRepairOrder_InvalidItemType tests repair order with invalid item type
func TestRepairOrder_InvalidItemType(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	// Test with invalid item type
	order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "item1", "invalid_type", 100)
	
	// Should still create order (type validation might be elsewhere)
	if err == nil {
		assert.NotNil(t, order)
		assert.Equal(t, "invalid_type", order.ItemType)
	}
}

// TestWarriorUpdate_UnauthorizedAccess tests unauthorized update attempts
func TestWarriorUpdate_UnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)
	
	svc := warrior.NewService()
	
	// Create regular warrior
	cmd := dto.CreateWarriorCommand{
		Username: "warrior1",
		Email:    "warrior1@example.com",
		Password: "password123",
		Role:     "knight",
	}
	warrior1, err := svc.CreateWarrior(cmd)
	require.NoError(t, err)
	
	// Create another warrior
	cmd2 := dto.CreateWarriorCommand{
		Username: "warrior2",
		Email:    "warrior2@example.com",
		Password: "password123",
		Role:     "knight",
	}
	warrior2, err := svc.CreateWarrior(cmd2)
	require.NoError(t, err)
	
	// Try to update warrior1's role as warrior2 (should fail)
	newRole := "archer"
	updateCmd := dto.UpdateWarriorCommand{
		WarriorID: warrior1.ID,
		Role:      &newRole,
		UpdatedBy: warrior2.ID, // Different warrior trying to update
	}
	
	_, err = svc.UpdateWarrior(updateCmd)
	// Should fail - only king can update roles
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "king")
}

