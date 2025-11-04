package repair

import "github.com/google/wire"

// ProviderSet provides DI for CQRS components of repair service
var ProviderSet = wire.NewSet(
    GetRepository,
    NewService,
)


