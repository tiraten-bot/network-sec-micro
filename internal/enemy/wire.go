package enemy

import "github.com/google/wire"

// ProviderSet is a Wire provider set for enemy service
var ProviderSet = wire.NewSet(
	NewService,
)

