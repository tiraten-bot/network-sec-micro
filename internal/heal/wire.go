package heal

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for heal service
var ProviderSet = wire.NewSet(
	NewService,
	NewHealServiceServer,
	GetRepository,
)

