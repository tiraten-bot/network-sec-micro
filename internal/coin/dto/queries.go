package dto

// GetBalanceQuery represents a query to get warrior balance
type GetBalanceQuery struct {
	WarriorID uint
}

// GetTransactionHistoryQuery represents a query to get transaction history
type GetTransactionHistoryQuery struct {
	WarriorID uint
	Limit     int
	Offset    int
}

