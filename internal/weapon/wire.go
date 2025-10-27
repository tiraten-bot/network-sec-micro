package weapon

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for weapon service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
)
