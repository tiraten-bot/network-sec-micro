package coin

import (
	"context"
	"errors"
	"fmt"

	"network-sec-micro/internal/coin/dto"

	"gorm.io/gorm"
)

// Service handles coin business logic with CQRS pattern and transaction safety
type Service struct {
	repo *Repository
}

// NewService creates a new coin service
func NewService() *Service {
	return &Service{
		repo: NewRepository(DB),
	}
}

// ==================== COMMANDS (WRITE OPERATIONS) ====================

// DeductCoins deducts coins from warrior's balance with transaction safety
func (s *Service) DeductCoins(ctx context.Context, cmd dto.DeductCoinsCommand) error {
	if cmd.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	var balanceBefore, balanceAfter int64
	var transaction *Transaction

	err := s.repo.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		// Get current balance with row lock
		balance, err := s.repo.GetWarriorBalance(ctx, cmd.WarriorID)
		if err != nil {
			return fmt.Errorf("failed to get warrior balance: %w", err)
		}

		balanceBefore = balance

		// Check sufficient balance
		if balanceBefore < cmd.Amount {
			return errors.New("insufficient balance")
		}

		// Calculate new balance
		balanceAfter = balanceBefore - cmd.Amount

		// Update warrior balance
		if err := s.repo.UpdateWarriorBalance(ctx, cmd.WarriorID, balanceAfter); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// Create transaction record
		transaction = &Transaction{
			WarriorID:       cmd.WarriorID,
			Amount:          -cmd.Amount,
			TransactionType: TransactionTypeDeduct,
			Reason:          cmd.Reason,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
		}

		if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("deduct coins failed: %w", err)
	}

	return nil
}

// AddCoins adds coins to warrior's balance with transaction safety
func (s *Service) AddCoins(ctx context.Context, cmd dto.AddCoinsCommand) error {
	if cmd.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	var balanceBefore, balanceAfter int64
	var transaction *Transaction

	err := s.repo.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		// Get current balance
		balance, err := s.repo.GetWarriorBalance(ctx, cmd.WarriorID)
		if err != nil {
			return fmt.Errorf("failed to get warrior balance: %w", err)
		}

		balanceBefore = balance
		balanceAfter = balanceBefore + cmd.Amount

		// Update warrior balance
		if err := s.repo.UpdateWarriorBalance(ctx, cmd.WarriorID, balanceAfter); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// Create transaction record
		transaction = &Transaction{
			WarriorID:       cmd.WarriorID,
			Amount:          cmd.Amount,
			TransactionType: TransactionTypeAdd,
			Reason:          cmd.Reason,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
		}

		if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("add coins failed: %w", err)
	}

	return nil
}

// TransferCoins transfers coins between warriors with atomic transaction
func (s *Service) TransferCoins(ctx context.Context, cmd dto.TransferCoinsCommand) error {
	if cmd.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if cmd.FromWarriorID == cmd.ToWarriorID {
		return errors.New("cannot transfer to self")
	}

	err := s.repo.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		// Deduct from sender
		if err := s.DeductCoins(ctx, dto.DeductCoinsCommand{
			WarriorID: cmd.FromWarriorID,
			Amount:    cmd.Amount,
			Reason:    "transfer_out: " + cmd.Reason,
		}); err != nil {
			return fmt.Errorf("failed to deduct from sender: %w", err)
		}

		// Add to receiver
		if err := s.AddCoins(ctx, dto.AddCoinsCommand{
			WarriorID: cmd.ToWarriorID,
			Amount:    cmd.Amount,
			Reason:    "transfer_in: " + cmd.Reason,
		}); err != nil {
			return fmt.Errorf("failed to add to receiver: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transfer coins failed: %w", err)
	}

	return nil
}

// CreateTransaction creates a new coin transaction record
func (s *Service) CreateTransaction(ctx context.Context, cmd dto.CreateTransactionCommand) (*Transaction, error) {
	txType := TransactionType(cmd.TransactionType)
	if txType != TransactionTypeAdd && txType != TransactionTypeDeduct &&
		txType != TransactionTypeTransferIn && txType != TransactionTypeTransferOut {
		return nil, errors.New("invalid transaction type")
	}

	transaction := &Transaction{
		WarriorID:       cmd.WarriorID,
		Amount:          cmd.Amount,
		TransactionType: txType,
		Reason:          cmd.Reason,
		BalanceBefore:   cmd.BalanceBefore,
		BalanceAfter:    cmd.BalanceAfter,
	}

	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("create transaction failed: %w", err)
	}

	return transaction, nil
}

// ==================== QUERIES (READ OPERATIONS) ====================

// GetBalance gets warrior's coin balance
func (s *Service) GetBalance(ctx context.Context, query dto.GetBalanceQuery) (int64, error) {
	balance, err := s.repo.GetWarriorBalance(ctx, query.WarriorID)
	if err != nil {
		return 0, fmt.Errorf("get balance failed: %w", err)
	}
	return balance, nil
}

// GetTransactionHistory gets transaction history for a warrior
func (s *Service) GetTransactionHistory(ctx context.Context, query dto.GetTransactionHistoryQuery) ([]Transaction, int64, error) {
	transactions, count, err := s.repo.GetTransactionHistory(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("get transaction history failed: %w", err)
	}
	return transactions, count, nil
}
