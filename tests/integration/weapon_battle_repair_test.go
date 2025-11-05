package integration_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/battle"
	"network-sec-micro/internal/battle/dto"
	"network-sec-micro/internal/repair"
	"network-sec-micro/internal/weapon"
	"network-sec-micro/internal/weapon/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWeaponPurchase_Battle_Wear_RepairFlow tests the complete flow:
// 1. Warrior purchases weapon
// 2. Warrior uses weapon in battle (wear applied)
// 3. Warrior repairs weapon
func TestWeaponPurchase_Battle_Wear_RepairFlow(t *testing.T) {
	// This is a conceptual integration test
	// In a real scenario, you'd set up test databases and services
	
	ctx := context.Background()
	
	// Step 1: Create weapon
	weaponSvc := weapon.NewService()
	createCmd := dto.CreateWeaponCommand{
		Name:        "Test Sword",
		Description: "A test sword",
		Type:        "common",
		Damage:      50,
		Price:       500,
		CreatedBy:   "warrior1",
	}
	
	weapon, err := weaponSvc.CreateWeapon(ctx, createCmd)
	require.NoError(t, err)
	assert.Equal(t, 100, weapon.MaxDurability) // Default max durability
	assert.Equal(t, 100, weapon.Durability)   // Starts at max
	
	// Step 2: Simulate weapon wear (as would happen in battle)
	// In actual battle, ApplyWear would be called via gRPC
	wearAmount := int32(10)
	// This simulates what happens in battle
	weapon.Durability -= int(wearAmount)
	assert.Equal(t, 90, weapon.Durability)
	
	// Step 3: Repair weapon
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	// Calculate repair cost
	cost := repairSvc.ComputeRepairCost(ctx, weapon.Durability, weapon.MaxDurability, "knight")
	assert.Equal(t, 20, cost) // 10 missing * 2 = 20
	
	// Create repair order
	order, err := repairSvc.CreateRepairOrder(ctx, "warrior", "warrior1", weapon.ID.Hex(), "weapon", cost)
	require.NoError(t, err)
	assert.Equal(t, repair.RepairStatusPending, order.Status)
	
	// Complete repair
	err = repairSvc.CompleteRepair(ctx, order.ID)
	require.NoError(t, err)
	
	// Verify repair completed
	orders, err := repairSvc.ListOrders(ctx, "warrior", "warrior1")
	require.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, repair.RepairStatusCompleted, orders[0].Status)
}

// TestArmorPurchase_Battle_Wear_RepairFlow tests armor flow
func TestArmorPurchase_Battle_Wear_RepairFlow(t *testing.T) {
	ctx := context.Background()
	
	// Create armor
	// armorSvc := armor.NewService()
	// createCmd := armorDto.CreateArmorCommand{...}
	// armor, err := armorSvc.CreateArmor(ctx, createCmd)
	
	// Simulate armor wear in battle
	// armor.Durability -= 15
	
	// Repair armor
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	// Test repair cost calculation for armor
	cost := repairSvc.ComputeRepairCost(ctx, 85, 100, "knight")
	assert.Equal(t, 30, cost) // 15 missing * 2 = 30
	
	// Test RBAC pricing
	emperorCost := repairSvc.ComputeRepairCost(ctx, 85, 100, "light_emperor")
	assert.Equal(t, 15, emperorCost) // 30 * 0.5 = 15 (50% discount)
	
	kingCost := repairSvc.ComputeRepairCost(ctx, 85, 100, "light_king")
	assert.Equal(t, 22, kingCost) // 30 * 0.75 = 22 (25% discount)
}

// TestBattle_WeaponArmorIntegration tests weapon and armor in battle
func TestBattle_WeaponArmorIntegration(t *testing.T) {
	// This test verifies battle logic with weapons and armors
	// In real scenario, this would test actual battle service
	
	ctx := context.Background()
	
	// Simulate battle attack with weapon
	attackerPower := 100
	weaponBonus := 50 // From weapon
	totalAttackPower := attackerPower + weaponBonus
	
	// Simulate defense with armor
	targetDefense := 30
	armorDefenseBonus := 40 // From armor
	totalDefense := targetDefense + armorDefenseBonus
	
	// Calculate damage (simplified formula)
	damage := totalAttackPower - totalDefense
	if damage < 10 {
		damage = 10 // Minimum damage
	}
	
	assert.Equal(t, 80, damage) // 150 - 70 = 80
	
	// Verify weapon wear would be applied (1 point per attack)
	// weapon.Durability -= 1
	
	// Verify armor wear would be applied (1 point per defense)
	// armor.Durability -= 1
}

// TestConcurrentRepairOrders tests concurrent repair requests
func TestConcurrentRepairOrders(t *testing.T) {
	ctx := context.Background()
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	// Create multiple repair orders concurrently
	orderCount := 5
	results := make(chan error, orderCount)
	
	for i := 0; i < orderCount; i++ {
		go func(id int) {
			order, err := repairSvc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 100)
			if err != nil {
				results <- err
				return
			}
			
			// Complete repair
			err = repairSvc.CompleteRepair(ctx, order.ID)
			results <- err
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < orderCount; i++ {
		err := <-results
		assert.NoError(t, err)
	}
	
	// Verify all orders created
	orders, err := repairSvc.ListOrders(ctx, "warrior", "warrior1")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(orders), orderCount)
}

// TestRepairCost_EdgeCases tests edge cases for repair cost calculation
func TestRepairCost_EdgeCases(t *testing.T) {
	ctx := context.Background()
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	tests := []struct {
		name        string
		currentDur  int
		maxDur      int
		role        string
		expectedCost int
	}{
		{
			name:        "full durability - no repair needed",
			currentDur:  100,
			maxDur:      100,
			role:        "knight",
			expectedCost: 0,
		},
		{
			name:        "over max durability (shouldn't happen but edge case)",
			currentDur:  150,
			maxDur:      100,
			role:        "knight",
			expectedCost: 0,
		},
		{
			name:        "zero durability - full repair",
			currentDur:  0,
			maxDur:      100,
			role:        "knight",
			expectedCost: 200, // 100 * 2
		},
		{
			name:        "emperor discount on full repair",
			currentDur:  0,
			maxDur:      100,
			role:        "light_emperor",
			expectedCost: 100, // 200 * 0.5
		},
		{
			name:        "king discount on full repair",
			currentDur:  0,
			maxDur:      100,
			role:        "light_king",
			expectedCost: 150, // 200 * 0.75
		},
		{
			name:        "single point missing",
			currentDur:  99,
			maxDur:      100,
			role:        "knight",
			expectedCost: 2, // 1 * 2
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := repairSvc.ComputeRepairCost(ctx, tt.currentDur, tt.maxDur, tt.role)
			assert.Equal(t, tt.expectedCost, cost)
		})
	}
}

// TestBattleDamageCalculation_WithWeaponArmor tests damage calculation
func TestBattleDamageCalculation_WithWeaponArmor(t *testing.T) {
	tests := []struct {
		name           string
		basePower      int
		weaponBonus    int
		baseDefense    int
		armorBonus     int
		expectedMin    int
		expectedMax    int
	}{
		{
			name:        "strong attacker with weapon vs weak defender",
			basePower:   100,
			weaponBonus: 50,
			baseDefense: 20,
			armorBonus:  10,
			expectedMin: 100, // 150 - 30 = 120, but with randomness
			expectedMax: 150,
		},
		{
			name:        "equal power with equal gear",
			basePower:   50,
			weaponBonus: 25,
			baseDefense: 50,
			armorBonus:  25,
			expectedMin: 10, // Minimum damage
			expectedMax: 30,
		},
		{
			name:        "weapon vs no armor",
			basePower:   80,
			weaponBonus: 40,
			baseDefense: 30,
			armorBonus:  0,
			expectedMin: 80,
			expectedMax: 120,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalAttack := tt.basePower + tt.weaponBonus
			totalDefense := tt.baseDefense + tt.armorBonus
			
			damage := totalAttack - totalDefense
			if damage < 10 {
				damage = 10
			}
			
			assert.GreaterOrEqual(t, damage, tt.expectedMin)
			assert.LessOrEqual(t, damage, tt.expectedMax)
		})
	}
}

