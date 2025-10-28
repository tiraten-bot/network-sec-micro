package coin

import "github.com/google/wire"

// ProviderSet is a Wire provider set for coin service
var ProviderSet = wire.NewSet(
	NewService,
	NewCoinServiceServer,
)

