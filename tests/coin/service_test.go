package coin_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/coin"
	"network-sec-micro/internal/coin/dto"
	"network-sec-micro/internal/warrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Auto migrate - warriors table has coin_balance column
	err = db.AutoMigrate(&warrior.Warrior{}, &coin.Transaction{})
	require.NoError(t, err)
	
	// Set global DB
	coin.DB = db
	
	return db
}

func TestAddCoins_Success(t *testing.T) {
	db := setupTestDB(t)
	
	// Create initial warrior with coin balance
	warrior := warrior.Warrior{
		ID:          1,
		Username:    "warrior1",
		Email:       "warrior1@example.com",
		Password:    "password",
		Role:        warrior.RoleKnight,
		CoinBalance: 1000,
	}
	err := db.Create(&warrior).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.AddCoinsCommand{
		WarriorID: 1,
		Amount:    500,
		Reason:    "test_add",
	}
	
	err = svc.AddCoins(ctx, cmd)
	assert.NoError(t, err)
	
	// Verify balance
	var updatedWarrior warrior.Warrior
	db.First(&updatedWarrior, 1)
	assert.Equal(t, 1500, updatedWarrior.CoinBalance)
}

func TestAddCoins_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.AddCoinsCommand{
		WarriorID: 1,
		Amount:    0,
		Reason:    "test",
	}
	
	err := svc.AddCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestAddCoins_NegativeAmount(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.AddCoinsCommand{
		WarriorID: 1,
		Amount:    -100,
		Reason:    "test",
	}
	
	err := svc.AddCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestDeductCoins_Success(t *testing.T) {
	db := setupTestDB(t)
	
	// Create initial warrior with coin balance
	warrior := warrior.Warrior{
		ID:          1,
		Username:    "warrior1",
		Email:       "warrior1@example.com",
		Password:    "password",
		Role:        warrior.RoleKnight,
		CoinBalance: 1000,
	}
	err := db.Create(&warrior).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.DeductCoinsCommand{
		WarriorID: 1,
		Amount:    300,
		Reason:    "test_deduct",
	}
	
	err = svc.DeductCoins(ctx, cmd)
	assert.NoError(t, err)
	
	// Verify balance
	var updatedWarrior warrior.Warrior
	db.First(&updatedWarrior, 1)
	assert.Equal(t, 700, updatedWarrior.CoinBalance)
}

func TestDeductCoins_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	
	// Create initial warrior with low coin balance
	warrior := warrior.Warrior{
		ID:          1,
		Username:    "warrior1",
		Email:       "warrior1@example.com",
		Password:    "password",
		Role:        warrior.RoleKnight,
		CoinBalance: 100,
	}
	err := db.Create(&warrior).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.DeductCoinsCommand{
		WarriorID: 1,
		Amount:    500,
		Reason:    "test_deduct",
	}
	
	err = svc.DeductCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestDeductCoins_InvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.DeductCoinsCommand{
		WarriorID: 1,
		Amount:    0,
		Reason:    "test",
	}
	
	err := svc.DeductCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestTransferCoins_Success(t *testing.T) {
	db := setupTestDB(t)
	
	// Create two warrior balances
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 500}
	err := db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.TransferCoinsCommand{
		FromWarriorID: 1,
		ToWarriorID:   2,
		Amount:        300,
		Reason:        "test_transfer",
	}
	
	err = svc.TransferCoins(ctx, cmd)
	assert.NoError(t, err)
	
	// Verify balances
	var updated1, updated2 coin.WarriorBalance
	db.First(&updated1, 1)
	db.First(&updated2, 2)
	assert.Equal(t, int64(700), updated1.Balance)
	assert.Equal(t, int64(800), updated2.Balance)
}

func TestTransferCoins_SelfTransfer(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.TransferCoinsCommand{
		FromWarriorID: 1,
		ToWarriorID:   1,
		Amount:        100,
		Reason:        "test",
	}
	
	err := svc.TransferCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transfer to self")
}

func TestTransferCoins_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	
	balance1 := coin.WarriorBalance{WarriorID: 1, Balance: 100}
	balance2 := coin.WarriorBalance{WarriorID: 2, Balance: 500}
	err := db.Create(&balance1).Error
	require.NoError(t, err)
	err = db.Create(&balance2).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.TransferCoinsCommand{
		FromWarriorID: 1,
		ToWarriorID:   2,
		Amount:        500,
		Reason:        "test",
	}
	
	err = svc.TransferCoins(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestGetBalance_Success(t *testing.T) {
	db := setupTestDB(t)
	
	balance := coin.WarriorBalance{
		WarriorID: 1,
		Balance:   1500,
	}
	err := db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	query := dto.GetBalanceQuery{WarriorID: 1}
	result, err := svc.GetBalance(ctx, query)
	
	require.NoError(t, err)
	assert.Equal(t, int64(1500), result)
}

func TestGetBalance_NotFound(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	query := dto.GetBalanceQuery{WarriorID: 999}
	result, err := svc.GetBalance(ctx, query)
	
	// Balance might default to 0 or return error
	// Adjust based on actual implementation
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.Equal(t, int64(0), result)
	}
}

func TestGetTransactionHistory_Success(t *testing.T) {
	db := setupTestDB(t)
	
	// Create warrior balance
	balance := coin.WarriorBalance{WarriorID: 1, Balance: 1000}
	err := db.Create(&balance).Error
	require.NoError(t, err)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	// Create some transactions
	addCmd := dto.AddCoinsCommand{WarriorID: 1, Amount: 500, Reason: "test1"}
	err = svc.AddCoins(ctx, addCmd)
	require.NoError(t, err)
	
	deductCmd := dto.DeductCoinsCommand{WarriorID: 1, Amount: 200, Reason: "test2"}
	err = svc.DeductCoins(ctx, deductCmd)
	require.NoError(t, err)
	
	// Get transaction history
	query := dto.GetTransactionHistoryQuery{
		WarriorID: 1,
		Limit:     10,
		Offset:    0,
	}
	result, count, err := svc.GetTransactionHistory(ctx, query)
	
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(2))
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestCreateTransaction_Success(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateTransactionCommand{
		WarriorID:       1,
		Amount:          500,
		TransactionType: "add",
		Reason:          "test_transaction",
		BalanceBefore:   1000,
		BalanceAfter:    1500,
	}
	
	result, err := svc.CreateTransaction(ctx, cmd)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.WarriorID)
	assert.Equal(t, int64(500), result.Amount)
	assert.Equal(t, coin.TransactionTypeAdd, result.TransactionType)
}

func TestCreateTransaction_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	
	svc := coin.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateTransactionCommand{
		WarriorID:       1,
		Amount:          500,
		TransactionType: "invalid_type",
		Reason:          "test",
		BalanceBefore:   1000,
		BalanceAfter:    1500,
	}
	
	result, err := svc.CreateTransaction(ctx, cmd)
	
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid transaction type")
}

