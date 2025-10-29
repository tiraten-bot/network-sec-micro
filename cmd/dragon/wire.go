//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/dragon"
)

// InitializeDragonApp initializes the dragon application using Wire
func InitializeDragonApp() (*dragon.Service, *dragon.Handler, error) {
	wire.Build(dragon.ProviderSet)
	return nil, nil, nil
}
