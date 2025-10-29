package coin

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/coin/dto"

	"gorm.io/gorm"
)

// Repository handles database operations with transaction safety
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetWarriorBalance gets warrior's balance with row lock for safety
func (r *Repository) GetWarriorBalance(ctx context.Context, warriorID uint) (int64, error) {
	var balance int64
	err := r.db.WithContext(ctx).
		Table("warriors").
		Select("coin_balance").
		Where("id = ?", warriorID).
		Row().Scan(&balance)
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("warrior not found")
		}
		return 0, fmt.Errorf("failed to get warrior balance: %w", err)
	}
	
	return balance, nil
}

// UpdateWarriorBalance updates warrior's balance with transaction safety
func (r *Repository) UpdateWarriorBalance(ctx context.Context, warriorID uint, newBalance int64) error {
	result := r.db.WithContext(ctx).
		Table("warriors").
		Where("id = ?", warriorID).
		Update("coin_balance", newBalance)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update warrior balance: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return errors.New("warrior not found")
	}
	
	return nil
}

// CreateTransaction creates a transaction record
func (r *Repository) CreateTransaction(ctx context.Context, tx *Transaction) error {
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

// GetTransactionHistory gets transaction history with pagination
func (r *Repository) GetTransactionHistory(ctx context.Context, query dto.GetTransactionHistoryQuery) ([]Transaction, int64, error) {
	var transactions []Transaction
	var count int64

	dbQuery := r.db.WithContext(ctx).Model(&Transaction{}).Where("warrior_id = ?", query.WarriorID)

	if err := dbQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	if query.Limit > 0 {
		dbQuery = dbQuery.Limit(query.Limit)
	}
	if query.Offset > 0 {
		dbQuery = dbQuery.Offset(query.Offset)
	}

	if err := dbQuery.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return transactions, count, nil
}

// ExecuteInTransaction executes multiple operations in a single transaction
func (r *Repository) ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
