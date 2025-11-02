//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/battlespell"
)

// InitializeApp initializes the application using Wire
func InitializeApp() (*battlespell.Service, *battlespell.Handler, error) {
	wire.Build(battlespell.ProviderSet)
	return nil, nil, nil
}

