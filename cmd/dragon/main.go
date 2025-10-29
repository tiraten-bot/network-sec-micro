package main

import (
	"log"
	"os"

	"network-sec-micro/internal/dragon"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	if err := dragon.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_HOST")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}

	if err := dragon.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Set Gin to release mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize service and handler manually (bypass Wire)
	service := dragon.NewService()
	handler := dragon.NewHandler(service)

	// Setup graceful shutdown
	defer func() {
		log.Println("Shutting down...")
		dragon.CloseKafkaPublisher()
	}()

	// Create Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	dragon.SetupRoutes(r, handler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Dragon service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
