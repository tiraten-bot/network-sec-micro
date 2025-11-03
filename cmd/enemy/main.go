package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"network-sec-micro/internal/enemy"
	pb "network-sec-micro/api/proto/enemy"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := enemy.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}

	if err := enemy.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Set Gin to release mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize service using Wire (or manual)
	// service, err := InitializeEnemyApp()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize app: %v", err)
	// }
	_ = enemy.NewService() // Service initialized
	
	// Setup graceful shutdown
	defer func() {
		log.Println("Shutting down...")
		enemy.CloseKafkaPublisher()
	}()

	log.Println("Enemy Service starting...")
	
	// Keep running (will add routes later)
	select {}
}

