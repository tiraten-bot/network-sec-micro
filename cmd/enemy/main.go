package main

import (
	"log"
	"os"

	"network-sec-micro/internal/enemy"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	if err := enemy.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Warrior gRPC client
ครั้ง warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
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

	// TODO: Add service and handler initialization
	log.Println("Enemy Service starting...")
	
	// Keep running (will add routes later)
	select {}
}

