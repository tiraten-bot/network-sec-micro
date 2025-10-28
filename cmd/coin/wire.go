//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/coin"
)

// InitializeCoinApp initializes the coin application using Wire
func InitializeCoinApp() (*coin.Service, *coin.CoinServiceServer, error) {
	wire.Build(coin.ProviderSet)
	return nil, nil, nil
}

