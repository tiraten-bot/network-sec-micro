//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/arena"
)

// InitializeApp initializes the application using Wire (CQRS pattern)
func InitializeApp() (*arena.Service, *arena.Handler, error) {
	wire.Build(arena.ProviderSet)
	return nil, nil, nil
}

