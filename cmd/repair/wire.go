//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "network-sec-micro/internal/repair"
)

// InitializeRepair constructs the repair service via Wire (CQRS/DI)
func InitializeRepair() (*repair.Service, error) {
    wire.Build(repair.ProviderSet)
    return nil, nil
}


