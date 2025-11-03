package arenaspell

import (
    "github.com/google/wire"
)

// ProviderSet is a Wire provider set for arenaspell service
var ProviderSet = wire.NewSet(
    NewService,
    NewHandler,
    NewArenaSpellServiceServer,
)


