package arena

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for arena service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
)

