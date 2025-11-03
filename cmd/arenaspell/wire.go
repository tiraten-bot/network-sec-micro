//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    arenaspell "network-sec-micro/internal/arenaspell"
)

// InitializeApp initializes the arenaspell application using Wire (CQRS pattern)
func InitializeApp() (*arenaspell.Service, *arenaspell.Handler, *arenaspell.ArenaSpellServiceServer, error) {
    wire.Build(arenaspell.ProviderSet)
    return nil, nil, nil, nil
}


