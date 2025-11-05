package coin_test

import (
	"network-sec-micro/internal/coin"
	"gorm.io/gorm"
)

// newTestService creates a coin service with a custom DB for testing
func newTestService(db *gorm.DB) *coin.Service {
	// Temporarily save old DB
	oldDB := coin.DB
	// Set test DB
	coin.DB = db
	// Create service (uses coin.DB)
	svc := coin.NewService()
	// Restore old DB (for other tests)
	coin.DB = oldDB
	// But service's repo still uses the test db we set
	return svc
}

