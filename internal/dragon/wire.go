package dragon

import "github.com/google/wire"

// ProviderSet is a Wire provider set for dragon service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
)
