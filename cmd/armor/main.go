// @title Armor Service API
// @version 1.0
// @description Armor service handles armor creation, purchase, and management. Purchases trigger Kafka events for coin deduction.
// @host localhost:8089
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
    "log"
    "net"
    "os"
    "sync"

    "network-sec-micro/internal/armor"
    pbArmor "network-sec-micro/api/proto/armor"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
)

func main() {
    if err := armor.InitDatabase(); err != nil { log.Fatalf("Failed to init db: %v", err) }

    service := armor.NewService()
    handler := armor.NewHandler(service)

    defer func(){ _ = armor.CloseKafkaPublisher() }()

    if os.Getenv("GIN_MODE") == "release" { gin.SetMode(gin.ReleaseMode) }
    r := gin.Default()
    // Setup routes (includes health check and metrics)
    armor.SetupRoutes(r, handler)

    httpPort := os.Getenv("PORT"); if httpPort == "" { httpPort = "8089" }
    grpcPort := os.Getenv("GRPC_PORT"); if grpcPort == "" { grpcPort = "50059" }

    var wg sync.WaitGroup
    wg.Add(1)
    go func(){
        defer wg.Done()
        log.Printf("Armor HTTP starting :%s", httpPort)
        if err := r.Run(":"+httpPort); err != nil { log.Fatalf("http error: %v", err) }
    }()

    wg.Add(1)
    go func(){
        defer wg.Done()
        lis, err := net.Listen("tcp", ":"+grpcPort); if err != nil { log.Fatalf("grpc listen: %v", err) }
        s := grpc.NewServer()
        pbArmor.RegisterArmorServiceServer(s, armor.NewArmorServiceServer())
        log.Printf("Armor gRPC starting :%s", grpcPort)
        if err := s.Serve(lis); err != nil { log.Fatalf("grpc serve: %v", err) }
    }()

    wg.Wait()
}


