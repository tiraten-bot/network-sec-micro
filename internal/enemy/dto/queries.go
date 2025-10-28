package dto

// GetEnemyQuery represents a query to get enemy by ID
type GetEnemyQuery struct {
	EnemyID string
}

// GetEnemiesByTypeQuery represents a query to get enemies by type
type GetEnemiesByTypeQuery struct {
	Type  string
	Limit int
	Offset int
}

// GetEnemiesByCreatorQuery represents a query to get enemies by creator
type GetEnemiesByCreatorQuery struct {
	CreatedBy string
	Limit     int
	Offset    int
}

