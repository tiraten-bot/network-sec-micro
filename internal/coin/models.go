package coin

import "time"

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeAdd       TransactionType = "add"
	TransactionTypeDeduct    TransactionType = "deduct"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
	TransactionTypeTransferOut TransactionType = "transfer_out"
)

// Transaction represents a coin transaction
type Transaction struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	WarriorID       uint            `gorm:"not null;index" json:"warrior_id"`
	Amount          int64           `gorm:"not null" json:"amount"` // positive for add, negative for deduct
	TransactionType TransactionType `gorm:"type:varchar(20);not null" json:"transaction_type"`
	Reason          string          `gorm:"type:text" json:"reason"`
	BalanceBefore   int64           `gorm:"not null" json:"balance_before"`
	BalanceAfter    int64           `gorm:"not null" json:"balance_after"`
	CreatedAt       time.Time       `json:"created_at"`
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "coin_transactions"
}

