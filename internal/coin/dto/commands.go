package dto

// DeductCoinsCommand represents a command to deduct coins
type DeductCoinsCommand struct {
	WarriorID uint
	Amount    int64
	Reason    string
}

// AddCoinsCommand represents a command to add coins
type AddCoinsCommand struct {
	WarriorID uint
	Amount    int64
	Reason    string
}

// TransferCoinsCommand represents a command to transfer coins
type TransferCoinsCommand struct {
	FromWarriorID uint
	ToWarriorID   uint
	Amount        int64
	Reason        string
}

// CreateTransactionCommand represents a command to create a transaction record
type CreateTransactionCommand struct {
	WarriorID       uint
	Amount          int64
	TransactionType string
	Reason          string
	BalanceBefore   int64
	BalanceAfter    int64
}

