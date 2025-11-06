package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/enemy"
	"network-sec-micro/internal/enemy"
	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"
	"network-sec-micro/pkg/secrets"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := enemy.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Warrior gRPC client
	warriorAddr := secrets.GetOrDefault("WARRIOR_GRPC_ADDR", "localhost:50052")

	if err := enemy.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Set Gin to release mode
	if secrets.GetOrDefault("GIN_MODE", "") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize service
	service := enemy.NewService()
	grpcServer := enemy.NewEnemyServiceServer(service)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start gRPC server
	grpcPort := secrets.GetOrDefault("GRPC_PORT", "50060")
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterEnemyServiceServer(grpcSrv, grpcServer)

	// Start metrics server
	metricsPort := secrets.GetOrDefault("METRICS_PORT", "8092")
	healthHandler := health.NewHandler(&health.MongoDBChecker{Client: enemy.Client, DBName: "mongodb"})
	go func() {
		if err := metrics.StartMetricsServer(metricsPort, healthHandler); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Enemy gRPC server starting on port %s", grpcPort)
		log.Printf("Enemy metrics server starting on port %s", metricsPort)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Println("Enemy Service starting...")

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down Enemy service...")
	grpcSrv.GracefulStop()
	enemy.CloseKafkaPublisher()
	log.Println("Enemy service stopped")
}
