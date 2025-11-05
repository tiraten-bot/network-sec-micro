package coin_test

import (
	"network-sec-micro/internal/coin"
	"gorm.io/gorm"
)

// newTestService creates a coin service with a custom DB for testing
func newTestService(db *gorm.DB) *coin.Service {
	// Save old DB
	oldDB := coin.DB
	// Set test DB
	coin.DB = db
	// Create service (uses coin.DB which is now set to our test DB)
	svc := coin.NewService()
	// Restore old DB immediately - service's repo already has reference to test DB
	coin.DB = oldDB
	return svc
}

