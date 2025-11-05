package complex_scenarios_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/battle"
	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/repair"

	"github.com/stretchr/testify/assert"
)

// TestCompleteWarriorFlow simulates complete warrior lifecycle
func TestCompleteWarriorFlow(t *testing.T) {
	// This test simulates:
	// 1. Warrior creates account
	// 2. Warrior purchases weapon
	// 3. Warrior purchases armor
	// 4. Warrior participates in battle
	// 5. Weapon and armor take wear
	// 6. Warrior repairs weapon and armor
	// 7. Warrior earns coins from battle
	// 8. Warrior purchases more items
	
	ctx := context.Background()
	
	// Step 1: Initial setup (would be done via warrior service)
	warriorID := "warrior1"
	initialCoins := int64(1000)
	
	// Step 2: Purchase weapon (would deduct coins via Kafka event)
	weaponPrice := 500
	coinsAfterWeapon := initialCoins - int64(weaponPrice)
	assert.Equal(t, int64(500), coinsAfterWeapon)
	
	// Step 3: Purchase armor
	armorPrice := 300
	coinsAfterArmor := coinsAfterWeapon - int64(armorPrice)
	assert.Equal(t, int64(200), coinsAfterArmor)
	
	// Step 4: Battle occurs (weapon and armor used)
	weaponDurabilityBefore := 100
	armorDurabilityBefore := 100
	
	// Simulate battle wear
	weaponWear := 5
	armorWear := 3
	
	weaponDurabilityAfter := weaponDurabilityBefore - weaponWear
	armorDurabilityAfter := armorDurabilityBefore - armorWear
	
	assert.Equal(t, 95, weaponDurabilityAfter)
	assert.Equal(t, 97, armorDurabilityAfter)
	
	// Step 5: Calculate repair costs
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	weaponRepairCost := repairSvc.ComputeRepairCost(ctx, weaponDurabilityAfter, weaponDurabilityBefore, "knight")
	armorRepairCost := repairSvc.ComputeRepairCost(ctx, armorDurabilityAfter, armorDurabilityBefore, "knight")
	
	assert.Equal(t, 10, weaponRepairCost) // (100-95) * 2 = 10
	assert.Equal(t, 6, armorRepairCost)   // (100-97) * 2 = 6
	
	// Step 6: Warrior earns coins from battle
	battleCoinsEarned := 200
	coinsAfterBattle := coinsAfterArmor + int64(battleCoinsEarned)
	assert.Equal(t, int64(400), coinsAfterBattle)
	
	// Step 7: Repair items
	totalRepairCost := weaponRepairCost + armorRepairCost
	coinsAfterRepair := coinsAfterBattle - int64(totalRepairCost)
	assert.Equal(t, int64(384), coinsAfterRepair) // 400 - 16
	
	// Verify warrior can afford repairs
	assert.GreaterOrEqual(t, coinsAfterRepair, int64(0))
}

// TestBattleWithWeaponArmorFlow tests complete battle flow with weapon and armor
func TestBattleWithWeaponArmorFlow(t *testing.T) {
	// Simulate battle with:
	// - Attacker has weapon
	// - Defender has armor
	// - Both take wear
	// - Battle concludes
	
	// Attacker setup
	attackerBasePower := 100
	attackerWeaponDamage := 50
	attackerTotalPower := attackerBasePower + attackerWeaponDamage
	
	// Defender setup
	defenderBaseDefense := 40
	defenderArmorDefense := 30
	defenderTotalDefense := defenderBaseDefense + defenderArmorDefense
	
	// Calculate damage
	damage := attackerTotalPower - defenderTotalDefense
	if damage < 10 {
		damage = 10
	}
	
	assert.Equal(t, 80, damage) // 150 - 70 = 80
	
	// Apply wear
	attackerWeaponDurability := 100
	attackerWeaponDurability -= 1 // Wear from using weapon
	assert.Equal(t, 99, attackerWeaponDurability)
	
	defenderArmorDurability := 100
	defenderArmorDurability -= 1 // Wear from defending
	assert.Equal(t, 99, defenderArmorDurability)
	
	// Apply damage
	defenderHP := 200
	defenderHP -= damage
	assert.Equal(t, 120, defenderHP)
	
	// Battle continues...
	for turn := 0; turn < 5; turn++ {
		damage := attackerTotalPower - defenderTotalDefense
		if damage < 10 {
			damage = 10
		}
		defenderHP -= damage
		attackerWeaponDurability -= 1
		defenderArmorDurability -= 1
		
		if defenderHP <= 0 {
			break
		}
	}
	
	// Verify wear accumulated
	assert.Less(t, attackerWeaponDurability, 100)
	assert.Less(t, defenderArmorDurability, 100)
	
	// Verify defender defeated or HP reduced
	assert.LessOrEqual(t, defenderHP, 0)
}

// TestRepairWithRBACFlow tests repair flow with different RBAC roles
func TestRepairWithRBACFlow(t *testing.T) {
	ctx := context.Background()
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	// Test repair costs for different roles
	weaponDurability := 50
	maxDurability := 100
	
	// Regular warrior
	warriorCost := repairSvc.ComputeRepairCost(ctx, weaponDurability, maxDurability, "knight")
	assert.Equal(t, 100, warriorCost) // (100-50) * 2 = 100
	
	// King (25% discount)
	kingCost := repairSvc.ComputeRepairCost(ctx, weaponDurability, maxDurability, "light_king")
	assert.Equal(t, 75, kingCost) // 100 * 0.75 = 75
	
	// Emperor (50% discount)
	emperorCost := repairSvc.ComputeRepairCost(ctx, weaponDurability, maxDurability, "light_emperor")
	assert.Equal(t, 50, emperorCost) // 100 * 0.5 = 50
	
	// Verify discounts are correct
	assert.Greater(t, warriorCost, kingCost)
	assert.Greater(t, kingCost, emperorCost)
	assert.Equal(t, warriorCost/2, emperorCost)
}

// TestMultipleBattles_WearAccumulation tests wear accumulation over multiple battles
func TestMultipleBattles_WearAccumulation(t *testing.T) {
	// Simulate warrior participating in multiple battles
	weaponDurability := 100
	armorDurability := 100
	
	battles := 10
	wearPerBattle := 3
	
	for i := 0; i < battles; i++ {
		weaponDurability -= wearPerBattle
		armorDurability -= wearPerBattle
		
		// Check if items are broken
		if weaponDurability <= 0 {
			weaponDurability = 0
			// Weapon is broken, cannot use
		}
		if armorDurability <= 0 {
			armorDurability = 0
			// Armor is broken, cannot use
		}
	}
	
	assert.Equal(t, 70, weaponDurability) // 100 - (10 * 3) = 70
	assert.Equal(t, 70, armorDurability)
	
	// Calculate repair cost
	ctx := context.Background()
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	weaponRepairCost := repairSvc.ComputeRepairCost(ctx, weaponDurability, 100, "knight")
	armorRepairCost := repairSvc.ComputeRepairCost(ctx, armorDurability, 100, "knight")
	
	assert.Equal(t, 60, weaponRepairCost) // (100-70) * 2 = 60
	assert.Equal(t, 60, armorRepairCost)
	
	totalRepairCost := weaponRepairCost + armorRepairCost
	assert.Equal(t, 120, totalRepairCost)
}

// TestBattleRewards_CoinCalculation tests coin rewards from battles
func TestBattleRewards_CoinCalculation(t *testing.T) {
	// Test battle reward calculation
	
	tests := []struct {
		name          string
		turns         int
		expectedCoins int
	}{
		{
			name:          "short battle",
			turns:         5,
			expectedCoins: 75, // 50 + (5 * 5)
		},
		{
			name:          "medium battle",
			turns:         10,
			expectedCoins: 100, // 50 + (10 * 5)
		},
		{
			name:          "long battle",
			turns:         20,
			expectedCoins: 150, // 50 + (20 * 5)
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseCoins := 50
			coinsPerTurn := 5
			totalCoins := baseCoins + (tt.turns * coinsPerTurn)
			
			assert.Equal(t, tt.expectedCoins, totalCoins)
		})
	}
}

// TestTeamBattle_ComplexScenario tests complex team battle scenario
func TestTeamBattle_ComplexScenario(t *testing.T) {
	// Simulate team battle with multiple participants
	// Each with weapons and armors
	
	lightSideParticipants := []struct {
		id       string
		weaponDM int
		armorDF  int
		hp       int
	}{
		{"warrior1", 50, 30, 200},
		{"warrior2", 40, 25, 180},
		{"warrior3", 35, 20, 150},
	}
	
	darkSideParticipants := []struct {
		id       string
		weaponDM int
		armorDF  int
		hp       int
	}{
		{"enemy1", 45, 28, 190},
		{"enemy2", 38, 22, 170},
	}
	
	// Simulate attacks
	turn := 0
	for {
		turn++
		
		// Light side attacks
		for _, attacker := range lightSideParticipants {
			if attacker.hp <= 0 {
				continue
			}
			
			// Find target
			target := &darkSideParticipants[0]
			if target.hp <= 0 {
				target = &darkSideParticipants[1]
			}
			
			// Calculate damage
			damage := attacker.weaponDM - target.armorDF
			if damage < 10 {
				damage = 10
			}
			
			target.hp -= damage
			
			// Apply wear
			attacker.weaponDM -= 1 // Simplified
			target.armorDF -= 1   // Simplified
		}
		
		// Dark side attacks
		for _, attacker := range darkSideParticipants {
			if attacker.hp <= 0 {
				continue
			}
			
			// Find target
			target := &lightSideParticipants[0]
			if target.hp <= 0 {
				target = &lightSideParticipants[1]
				if target.hp <= 0 {
					target = &lightSideParticipants[2]
				}
			}
			
			// Calculate damage
			damage := attacker.weaponDM - target.armorDF
			if damage < 10 {
				damage = 10
			}
			
			target.hp -= damage
			
			// Apply wear
			attacker.weaponDM -= 1
			target.armorDF -= 1
		}
		
		// Check if battle is complete
		lightAlive := 0
		for _, p := range lightSideParticipants {
			if p.hp > 0 {
				lightAlive++
			}
		}
		
		darkAlive := 0
		for _, p := range darkSideParticipants {
			if p.hp > 0 {
				darkAlive++
			}
		}
		
		if lightAlive == 0 || darkAlive == 0 || turn >= 50 {
			break
		}
	}
	
	// Verify battle completed
	assert.Less(t, turn, 50)
	
	// Verify wear was applied
	for _, p := range lightSideParticipants {
		assert.Less(t, p.weaponDM, 50) // Should have decreased
	}
	
	for _, p := range darkSideParticipants {
		assert.Less(t, p.weaponDM, 45) // Should have decreased
	}
}

// TestEconomyFlow tests complete economy flow
func TestEconomyFlow(t *testing.T) {
	// Test complete economy:
	// 1. Warriors earn coins
	// 2. Warriors purchase items
	// 3. Items wear in battles
	// 4. Warriors repair items
	// 5. Cycle repeats
	
	warriorCoins := int64(1000)
	
	// Purchase weapon
	weaponPrice := 500
	warriorCoins -= int64(weaponPrice)
	
	// Purchase armor
	armorPrice := 300
	warriorCoins -= int64(armorPrice)
	
	assert.Equal(t, int64(200), warriorCoins)
	
	// Battle and earn coins
	battleReward := 150
	warriorCoins += int64(battleReward)
	
	assert.Equal(t, int64(350), warriorCoins)
	
	// Repair items
	ctx := context.Background()
	repairRepo := repair.GetRepository()
	repairSvc := repair.NewService(repairRepo)
	
	repairCost := repairSvc.ComputeRepairCost(ctx, 90, 100, "knight")
	warriorCoins -= int64(repairCost)
	
	assert.Equal(t, int64(330), warriorCoins) // 350 - 20
	
	// Verify economy is sustainable
	assert.Greater(t, warriorCoins, int64(0))
}

