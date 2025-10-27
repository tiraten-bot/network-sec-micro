package coin

import "fmt"

// Service handles coin business logic
type Service struct{}

// NewService creates a new coin service
func NewService() *Service {
	return &Service{}
}

// CreateTransaction creates a new coin transaction
func (s *Service) CreateTransaction(warriorID uint, amount int64, txType TransactionType, reason string, balanceBefore, balanceAfter int64) (*Transaction, error) {
	transaction := Transaction{
		WarriorID:       warriorID,
		Amount:          amount,
		TransactionType: txType,
		Reason:          reason,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
	}

	if err := DB.Create(&transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &transaction, nil
}

// GetTransactionsByWarrior gets transaction history for a warrior
func (s *Service) GetTransactionsByWarrior(warriorID uint, limit, offset int) ([]Transaction, int64, error) {
	var transactions []Transaction
	var count int64

	query := DB.Model(&Transaction{}).Where("warrior_id = ?", warriorID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return transactions, count, nil
}

