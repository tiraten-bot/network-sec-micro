package main

import (
    "log"
    "os"

    arenaspell "network-sec-micro/internal/arenaspell"

    "github.com/gin-gonic/gin"
)

func main() {
    if err := arenaspell.InitDatabase(); err != nil {
        log.Fatalf("arenaspell db init failed: %v", err)
    }

    svc := arenaspell.NewService()
    h := arenaspell.NewHandler(svc)

    r := gin.Default()
    arenaspell.SetupRoutes(r, h)

    addr := os.Getenv("ARENASPELL_HTTP_ADDR")
    if addr == "" {
        addr = ":8088"
    }
    if err := r.Run(addr); err != nil {
        log.Fatalf("arenaspell http failed: %v", err)
    }
}


