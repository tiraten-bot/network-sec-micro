//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/heal"
)

// InitializeApp initializes the heal application using Wire (CQRS pattern)
func InitializeApp() (*heal.Service, *heal.HealServiceServer, error) {
	wire.Build(heal.ProviderSet)
	return nil, nil, nil
}

