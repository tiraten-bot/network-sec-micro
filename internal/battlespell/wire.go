package battlespell

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for battlespell service
var ProviderSet = wire.NewSet(
	NewService,
	NewHandler,
	NewBattleSpellServiceServer,
)

