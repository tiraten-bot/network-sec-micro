//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "network-sec-micro/internal/weapon"
)

// InitializeWeaponApp initializes the weapon app using Wire
func InitializeWeaponApp() (*weapon.Service, *weapon.Handler, error) {
    wire.Build(weapon.ProviderSet)
    return nil, nil, nil
}


