package dto

// GetArmorsQuery represents a query to get armors
type GetArmorsQuery struct {
	Type      string
	CreatedBy string
	OwnedBy   string
}

