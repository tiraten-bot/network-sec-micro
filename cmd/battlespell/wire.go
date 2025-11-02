//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/battlespell"
)

// InitializeApp initializes the application using Wire (CQRS pattern)
func InitializeApp() (*battlespell.Service, *battlespell.Handler, *battlespell.BattleSpellServiceServer, error) {
	wire.Build(battlespell.ProviderSet)
	return nil, nil, nil, nil
}

