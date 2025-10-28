package coin

import (
	"errors"
	"fmt"

	"network-sec-micro/internal/coin/dto"

	"gorm.io/gorm"
)

// Service handles coin business logic with CQRS pattern
type Service struct{}

// NewService creates a new coin service
func NewService() *Service {
	return &Service{}
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// DeductCoins deducts coins from warrior's balance
func (s *Service) DeductCoins(cmd dto.DeductCoinsCommand) error {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", cmd.WarriorID).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("warrior not found")
		}
		return err
	}

	balanceBefore := int64(warrior.CoinBalance)

	if balanceBefore < cmd.Amount {
		return errors.New("insufficient balance")
	}

	warrior.CoinBalance -= int(cmd.Amount)
	balanceAfter := int64(warrior.CoinBalance)

	if err := DB.Table("warriors").Where("id = ?", cmd.WarriorID).Update("coin_balance", warrior.CoinBalance).Error; err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create transaction record
	if _, err := s.CreateTransaction(dto.CreateTransactionCommand{
		WarriorID:       cmd.WarriorID,
		Amount:          -cmd.Amount,
		TransactionType: string(TransactionTypeDeduct),
		Reason:          cmd.Reason,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
	}); err != nil {
		// Log error but don't fail the deduction
		fmt.Printf("Failed to create transaction record: %v", err)
	}

	return nil
}

// AddCoins adds coins to warrior's balance
func (s *Service) AddCoins(cmd dto.AddCoinsCommand) error {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", cmd.WarriorID).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("warrior not found")
		}
		return err
	}

	balanceBefore := int64(warrior.CoinBalance)
	warrior.CoinBalance += int(cmd.Amount)
	balanceAfter := int64(warrior.CoinBalance)

	if err := DB.Table("warriors").Where("id = ?", cmd.WarriorID).Update("coin_balance", warrior.CoinBalance).Error; err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create transaction record
	if _, err := s.CreateTransaction(dto.CreateTransactionCommand{
		WarriorID:       cmd.WarriorID,
		Amount:          cmd.Amount,
		TransactionType: string(TransactionTypeAdd),
		Reason:          cmd.Reason,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
	}); err != nil {
		// Log error but don't fail the addition
		fmt.Printf("Failed to create transaction record: %v", err)
	}

	return nil
}

// TransferCoins transfers coins between warriors
func (s *Service) TransferCoins(cmd dto.TransferCoinsCommand) error {
	// Deduct from sender
	if err := s.DeductCoins(dto.DeductCoinsCommand{
		WarriorID: cmd.FromWarriorID,
		Amount:    cmd.Amount,
		Reason:    "transfer_out: " + cmd.Reason,
	}); err != nil {
		return fmt.Errorf("failed to deduct from sender: %w", err)
	}

	// Add to receiver
	if err := s.AddCoins(dto.AddCoinsCommand{
		WarriorID: cmd.ToWarriorID,
		Amount:    cmd.Amount,
		Reason:    "transfer_in: " + cmd.Reason,
	}); err != nil {
		// Rollback: add coins back to sender
		s.AddCoins(dto.AddCoinsCommand{
			WarriorID: cmd.FromWarriorID,
			Amount:    cmd.Amount,
			Reason:    "rollback transfer_out",
		})
		return fmt.Errorf("failed to add to receiver: %w", err)
	}

	return nil
}

// CreateTransaction creates a new coin transaction record
func (s *Service) CreateTransaction(cmd dto.CreateTransactionCommand) (*Transaction, error) {
	txType := TransactionType(cmd.TransactionType)
	if txType != TransactionTypeAdd && txType != TransactionTypeDeduct &&
		txType != TransactionTypeTransferIn && txType != TransactionTypeTransferOut {
		return nil, errors.New("invalid transaction type")
	}

	transaction := Transaction{
		WarriorID:       cmd.WarriorID,
		Amount:          cmd.Amount,
		TransactionType: txType,
		Reason:          cmd.Reason,
		BalanceBefore:   cmd.BalanceBefore,
		BalanceAfter:    cmd.BalanceAfter,
	}

	if err := DB.Create(&transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &transaction, nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetBalance gets warrior's coin balance
func (s *Service) GetBalance(query dto.GetBalanceQuery) (int64, error) {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", query.WarriorID).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("warrior not found")
		}
		return 0, err
	}

	return int64(warrior.CoinBalance), nil
}

// GetTransactionHistory gets transaction history for a warrior
func (s *Service) GetTransactionHistory(query dto.GetTransactionHistoryQuery) ([]Transaction, int64, error) {
	var transactions []Transaction
	var count int64

	dbQuery := DB.Model(&Transaction{}).Where("warrior_id = ?", query.WarriorID)

	if err := dbQuery.Count(&count).Error; err != nil {
		return nil, 0, err
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
