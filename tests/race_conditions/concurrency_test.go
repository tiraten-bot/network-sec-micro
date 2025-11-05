package race_conditions_test

import (
	"context"
	"sync"
	"testing"

	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"
	"network-sec-micro/internal/repair"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCoinDeduction_RaceCondition tests race condition in coin deduction
func TestCoinDeduction_RaceCondition(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create warrior with balance
	balance := coin.WarriorBalance{
		WarriorID: 1,
		Balance:   1000,
	}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Multiple goroutines trying to deduct same amount
	concurrency := 10
	deductAmount := int64(100) // Each tries to deduct 100
	
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			cmd := dto.DeductCoinsCommand{
				WarriorID: 1,
				Amount:    deductAmount,
				Reason:    "race_test",
			}
			
			err := svc.DeductCoins(ctx, cmd)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// Verify only some succeeded (not all)
	assert.Greater(t, successCount, 0, "Some deductions should succeed")
	assert.Less(t, successCount, concurrency, "Not all should succeed due to insufficient balance")
	
	// Verify final balance
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	// Should have deducted exactly successCount * deductAmount
	expectedBalance := int64(1000) - (int64(successCount) * deductAmount)
	assert.Equal(t, expectedBalance, finalBalance.Balance)
}

// TestRepairOrderCreation_RaceCondition tests race condition in repair order creation
func TestRepairOrderCreation_RaceCondition(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	concurrency := 50
	
	var wg sync.WaitGroup
	orderIDs := make(chan uint, concurrency)
	var mu sync.Mutex
	orderCount := 0
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 100)
			if err != nil {
				return
			}
			
			mu.Lock()
			orderCount++
			mu.Unlock()
			
			orderIDs <- order.ID
		}()
	}
	
	wg.Wait()
	close(orderIDs)
	
	// Verify all orders created
	assert.Equal(t, concurrency, orderCount)
	
	// Verify all orders are unique
	uniqueIDs := make(map[uint]bool)
	for id := range orderIDs {
		assert.False(t, uniqueIDs[id], "Order ID should be unique: %d", id)
		uniqueIDs[id] = true
	}
}

// TestCoinTransfer_RaceCondition tests race condition in coin transfers
func TestCoinTransfer_RaceCondition(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create two warriors
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 1000}
	err = db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Multiple goroutines transferring between same two warriors
	concurrency := 20
	transferAmount := int64(50)
	
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			cmd := dto.TransferCoinsCommand{
				FromWarriorID: 1,
				ToWarriorID:   2,
				Amount:        transferAmount,
				Reason:        "race_test",
			}
			
			err := svc.TransferCoins(ctx, cmd)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	// Verify transfers succeeded
	assert.Greater(t, successCount, 0)
	
	// Verify final balances
	var finalBalance1, finalBalance2 coin.WarriorBalance
	db.First(&finalBalance1, 1)
	db.First(&finalBalance2, 2)
	
	// Verify balance consistency
	expectedBalance1 := int64(1000) - (int64(successCount) * transferAmount)
	expectedBalance2 := int64(1000) + (int64(successCount) * transferAmount)
	
	assert.Equal(t, expectedBalance1, finalBalance1.Balance)
	assert.Equal(t, expectedBalance2, finalBalance2.Balance)
	
	// Verify total balance preserved
	totalBalance := finalBalance1.Balance + finalBalance2.Balance
	assert.Equal(t, int64(2000), totalBalance)
}

// TestConcurrentRepairCostCalculation tests concurrent repair cost calculations
func TestConcurrentRepairCostCalculation(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	concurrency := 100
	calculationsPerGoroutine := 100
	
	var wg sync.WaitGroup
	results := make(chan int, concurrency*calculationsPerGoroutine)
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < calculationsPerGoroutine; j++ {
				cost := svc.ComputeRepairCost(ctx, 50, 100, "knight")
				results <- cost
			}
		}(i)
	}
	
	wg.Wait()
	close(results)
	
	// Verify all results are correct
	resultCount := 0
	for cost := range results {
		resultCount++
		assert.Equal(t, 100, cost) // (100-50) * 2 = 100
	}
	
	assert.Equal(t, concurrency*calculationsPerGoroutine, resultCount)
}

// TestReadWriteRaceCondition tests read-write race condition
func TestReadWriteRaceCondition(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{})
	require.NoError(t, err)
	
	coin.DB = db
	
	balance := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Readers and writers concurrently
	readers := 10
	writers := 5
	
	var wg sync.WaitGroup
	readErrors := make(chan error, readers*100)
	writeErrors := make(chan error, writers*100)
	
	// Start readers
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				query := dto.GetBalanceQuery{WarriorID: 1}
				_, err := svc.GetBalance(ctx, query)
				if err != nil {
					readErrors <- err
				}
			}
		}()
	}
	
	// Start writers
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < 100; j++ {
				cmd := dto.AddCoinsCommand{
					WarriorID: 1,
					Amount:    1,
					Reason:    "race_test",
				}
				err := svc.AddCoins(ctx, cmd)
				if err != nil {
					writeErrors <- err
				}
			}
		}()
	}
	
	wg.Wait()
	close(readErrors)
	close(writeErrors)
	
	// Verify no errors
	readErrorCount := 0
	for err := range readErrors {
		if err != nil {
			readErrorCount++
		}
	}
	
	writeErrorCount := 0
	for err := range writeErrors {
		if err != nil {
			writeErrorCount++
		}
	}
	
	assert.Equal(t, 0, readErrorCount)
	assert.Equal(t, 0, writeErrorCount)
	
	// Verify final balance
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	expectedBalance := int64(1000) + int64(writers*100)
	assert.Equal(t, expectedBalance, finalBalance.Balance)
}

