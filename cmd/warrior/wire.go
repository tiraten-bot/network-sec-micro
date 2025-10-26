//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/warrior"
)

// InitializeApp initializes the application using Wire
func InitializeApp() (*warrior.Service, *warrior.Handler, error) {
	wire.Build(warrior.ProviderSet)
	return nil, nil, nil
}
