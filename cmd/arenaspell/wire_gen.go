package main

import (
    arenaspell "network-sec-micro/internal/arenaspell"
)

// InitializeApp is a fallback manual initializer if wire generation isn't run
func InitializeApp() (*arenaspell.Service, *arenaspell.Handler, *arenaspell.ArenaSpellServiceServer, error) {
    svc := arenaspell.NewService()
    h := arenaspell.NewHandler(svc)
    grpcSrv := arenaspell.NewArenaSpellServiceServer(svc)
    return svc, h, grpcSrv, nil
}


