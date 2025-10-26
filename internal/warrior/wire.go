package warrior

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for warrior service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
)

// InitializeService initializes the warrior service
func InitializeService() (*Service, *Handler, error) {
	// Initialize database first
	if err := InitDatabase(); err != nil {
		return nil, nil, err
	}

	service := NewService()
	handler := NewHandler()
	
	// Inject service into handler
	handler.service = service

	return service, handler, nil
}
