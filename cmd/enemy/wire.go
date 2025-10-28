//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"network-sec-micro/internal/enemy"
)

// InitializeEnemyApp initializes the enemy application using Wire
func InitializeEnemyApp() (*enemy.Service, error) {
	wire.Build(enemy.ProviderSet)
	return nil, nil
}

