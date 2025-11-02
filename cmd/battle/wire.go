//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/battle"
)

// InitializeApp initializes the application using Wire
func InitializeApp() (*battle.Service, *battle.Handler, error) {
	wire.Build(battle.ProviderSet)
	return nil, nil, nil
}

