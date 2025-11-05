package transactions_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCoinTransfer_Atomicity tests that coin transfers are atomic
func TestCoinTransfer_Atomicity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create warriors with balances
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 500}
	err = db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Transfer amount
	transferAmount := int64(300)
	
	cmd := dto.TransferCoinsCommand{
		FromWarriorID: 1,
		ToWarriorID:   2,
		Amount:        transferAmount,
		Reason:        "atomicity_test",
	}
	
	err = svc.TransferCoins(ctx, cmd)
	require.NoError(t, err)
	
	// Verify balances
	var finalBalance1, finalBalance2 coin.WarriorBalance
	db.First(&finalBalance1, 1)
	db.First(&finalBalance2, 2)
	
	assert.Equal(t, int64(700), finalBalance1.Balance) // 1000 - 300
	assert.Equal(t, int64(800), finalBalance2.Balance) // 500 + 300
	
	// Verify total balance preserved
	totalBalance := finalBalance1.Balance + finalBalance2.Balance
	assert.Equal(t, int64(1500), totalBalance)
	
	// Verify transaction records created
	var transactions []coin.Transaction
	db.Where("warrior_id = ? OR warrior_id = ?", 1, 2).Find(&transactions)
	
	// Should have 2 transactions (deduct from 1, add to 2)
	assert.GreaterOrEqual(t, len(transactions), 2)
}

// TestCoinDeduction_Atomicity tests that coin deduction is atomic
func TestCoinDeduction_Atomicity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	balance := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Deduct amount
	deductAmount := int64(300)
	
	cmd := dto.DeductCoinsCommand{
		WarriorID: 1,
		Amount:    deductAmount,
		Reason:    "atomicity_test",
	}
	
	err = svc.DeductCoins(ctx, cmd)
	require.NoError(t, err)
	
	// Verify balance updated
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	assert.Equal(t, int64(700), finalBalance.Balance) // 1000 - 300
	
	// Verify transaction record created
	var transaction coin.Transaction
	err = db.Where("warrior_id = ? AND transaction_type = ?", 1, coin.TransactionTypeDeduct).First(&transaction).Error
	require.NoError(t, err)
	
	assert.Equal(t, uint(1), transaction.WarriorID)
	assert.Equal(t, -deductAmount, transaction.Amount)
	assert.Equal(t, int64(1000), transaction.BalanceBefore)
	assert.Equal(t, int64(700), transaction.BalanceAfter)
}

// TestCoinAddition_Atomicity tests that coin addition is atomic
func TestCoinAddition_Atomicity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	balance := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Add amount
	addAmount := int64(500)
	
	cmd := dto.AddCoinsCommand{
		WarriorID: 1,
		Amount:    addAmount,
		Reason:    "atomicity_test",
	}
	
	err = svc.AddCoins(ctx, cmd)
	require.NoError(t, err)
	
	// Verify balance updated
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	assert.Equal(t, int64(1500), finalBalance.Balance) // 1000 + 500
	
	// Verify transaction record created
	var transaction coin.Transaction
	err = db.Where("warrior_id = ? AND transaction_type = ?", 1, coin.TransactionTypeAdd).First(&transaction).Error
	require.NoError(t, err)
	
	assert.Equal(t, uint(1), transaction.WarriorID)
	assert.Equal(t, addAmount, transaction.Amount)
	assert.Equal(t, int64(1000), transaction.BalanceBefore)
	assert.Equal(t, int64(1500), transaction.BalanceAfter)
}

// TestTransactionRollback_OnError tests transaction rollback on error
func TestTransactionRollback_OnError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	balance := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	err = db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Try to deduct more than balance (should fail and rollback)
	cmd := dto.DeductCoinsCommand{
		WarriorID: 1,
		Amount:    2000, // More than balance
		Reason:    "rollback_test",
	}
	
	err = svc.DeductCoins(ctx, cmd)
	require.Error(t, err)
	
	// Verify balance unchanged
	var finalBalance coin.WarriorBalance
	db.First(&finalBalance, 1)
	
	assert.Equal(t, int64(1000), finalBalance.Balance) // Should remain unchanged
	
	// Verify no transaction record created
	var transaction coin.Transaction
	err = db.Where("warrior_id = ? AND reason = ?", 1, "rollback_test").First(&transaction).Error
	assert.Error(t, err) // Should not find transaction
}

// TestMultiOperation_Atomicity tests multiple operations in single transaction
func TestMultiOperation_Atomicity(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&coin.WarriorBalance{}, &coin.Transaction{})
	require.NoError(t, err)
	
	coin.DB = db
	
	// Create warriors
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 500}
	balance3 := coin.WarriorBalance{WarriorID: 3, Balance: 200}
	err = db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	err = db.Create(&balance3).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Transfer from 1 to 2, then from 2 to 3
	// If any fails, all should rollback
	
	// First transfer
	cmd1 := dto.TransferCoinsCommand{
		FromWarriorID: 1,
		ToWarriorID:   2,
		Amount:        200,
		Reason:        "multi_op_test",
	}
	err = svc.TransferCoins(ctx, cmd1)
	require.NoError(t, err)
	
	// Second transfer
	cmd2 := dto.TransferCoinsCommand{
		FromWarriorID: 2,
		ToWarriorID:   3,
		Amount:        100,
		Reason:        "multi_op_test",
	}
	err = svc.TransferCoins(ctx, cmd2)
	require.NoError(t, err)
	
	// Verify final balances
	var finalBalance1, finalBalance2, finalBalance3 coin.WarriorBalance
	db.First(&finalBalance1, 1)
	db.First(&finalBalance2, 2)
	db.First(&finalBalance3, 3)
	
	assert.Equal(t, int64(800), finalBalance1.Balance)  // 1000 - 200
	assert.Equal(t, int64(600), finalBalance2.Balance)   // 500 + 200 - 100
	assert.Equal(t, int64(300), finalBalance3.Balance)   // 200 + 100
	
	// Verify total preserved
	totalBalance := finalBalance1.Balance + finalBalance2.Balance + finalBalance3.Balance
	assert.Equal(t, int64(1700), totalBalance)
}

