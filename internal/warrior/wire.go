package warrior

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for warrior service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
)

