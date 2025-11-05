package performance_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"
	"network-sec-micro/internal/repair"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestConcurrentRepairOrders_Load tests concurrent repair order creation
func TestConcurrentRepairOrders_Load(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	concurrency := 100
	ordersPerGoroutine := 10
	totalOrders := concurrency * ordersPerGoroutine
	
	var wg sync.WaitGroup
	errors := make(chan error, totalOrders)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < ordersPerGoroutine; j++ {
				order, err := svc.CreateRepairOrder(ctx, "warrior", "warrior1", "weapon1", "weapon", 100)
				if err != nil {
					errors <- err
					return
				}
				
				// Complete repair
				err = svc.CompleteRepair(ctx, order.ID)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	duration := time.Since(start)
	
	// Check for errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Logf("Error: %v", err)
		}
	}
	
	assert.Equal(t, 0, errorCount, "No errors expected in concurrent operations")
	assert.Less(t, duration, 30*time.Second, "Should complete within 30 seconds")
	
	t.Logf("Created %d orders in %v (%.2f ops/sec)", totalOrders, duration, float64(totalOrders)/duration.Seconds())
}

// TestCoinOperations_Load tests concurrent coin operations
func TestCoinOperations_Load(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create warrior balance
	balance := coin.WarriorBalance{
		WarriorID: 1,
		Balance:   1000000, // Large balance for load test
	}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	concurrency := 50
	operationsPerGoroutine := 20
	
	var wg sync.WaitGroup
	errors := make(chan error, concurrency*operationsPerGoroutine)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				// Alternating add and deduct
				if j%2 == 0 {
					cmd := dto.AddCoinsCommand{
						WarriorID: 1,
						Amount:    10,
						Reason:    "load_test",
					}
					err := svc.AddCoins(ctx, cmd)
					if err != nil {
						errors <- err
					}
				} else {
					cmd := dto.DeductCoinsCommand{
						WarriorID: 1,
						Amount:    5,
						Reason:    "load_test",
					}
					err := svc.DeductCoins(ctx, cmd)
					if err != nil {
						errors <- err
					}
				}
			}
		}()
	}
	
	wg.Wait()
	close(errors)
	
	duration := time.Since(start)
	
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}
	
	assert.Equal(t, 0, errorCount)
	assert.Less(t, duration, 10*time.Second)
	
	// Verify final balance
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	// Expected: 1000000 + (50 * 10 * 10) - (50 * 5 * 10) = 1000000 + 5000 - 2500 = 1002500
	expectedBalance := int64(1000000 + (concurrency * 10 * operationsPerGoroutine / 2) - (concurrency * 5 * operationsPerGoroutine / 2))
	assert.GreaterOrEqual(t, finalBalance.Balance, expectedBalance-100) // Allow small variance
}

// TestRepairCostCalculation_Performance tests repair cost calculation performance
func TestRepairCostCalculation_Performance(t *testing.T) {
	repo := repair.GetRepository()
	svc := repair.NewService(repo)
	ctx := context.Background()
	
	iterations := 10000
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_ = svc.ComputeRepairCost(ctx, 50, 100, "knight")
		_ = svc.ComputeRepairCost(ctx, 0, 1000, "light_emperor")
		_ = svc.ComputeRepairCost(ctx, 25, 100, "light_king")
	}
	
	duration := time.Since(start)
	
	avgTime := duration / time.Duration(iterations*3)
	
	t.Logf("Computed %d repair costs in %v (avg: %v per calculation)", iterations*3, duration, avgTime)
	assert.Less(t, avgTime, 1*time.Millisecond, "Each calculation should be very fast")
}

// TestConcurrentCoinTransfer_Load tests concurrent coin transfers
func TestConcurrentCoinTransfer_Load(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create multiple warriors
	warriors := 10
	for i := 1; i <= warriors; i++ {
		balance := coin.WarriorBalance{
			WarriorID: uint(i),
			Balance:   10000,
		}
		err := db.Create(&balance).Error
		require.NoError(t, err)
	}
	
	svc := coin.NewService()
	ctx := context.Background()
	
	concurrency := 20
	transfersPerGoroutine := 5
	
	var wg sync.WaitGroup
	errors := make(chan error, concurrency*transfersPerGoroutine)
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < transfersPerGoroutine; j++ {
				from := (id % warriors) + 1
				to := ((id + 1) % warriors) + 1
				if to == from {
					to = (to % warriors) + 1
				}
				
				cmd := dto.TransferCoinsCommand{
					FromWarriorID: uint(from),
					ToWarriorID:   uint(to),
					Amount:        10,
					Reason:        "load_test",
				}
				
				err := svc.TransferCoins(ctx, cmd)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	duration := time.Since(start)
	
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}
	
	assert.Equal(t, 0, errorCount)
	assert.Less(t, duration, 15*time.Second)
	
	t.Logf("Completed %d transfers in %v", concurrency*transfersPerGoroutine, duration)
}

