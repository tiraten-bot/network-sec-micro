package main

import (
	"context"
	"log"
	"os"

	"network-sec-micro/internal/arena"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	if err := arena.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}

	if err := arena.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Set Gin to release mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize service and handler using Wire
	service, handler, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app with Wire: %v", err)
	}

	// Initialize Kafka consumer
	if err := arena.InitKafkaConsumer(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka consumer: %v", err)
	} else {
		// Start consuming battle completed events
		consumer := arena.NewArenaConsumer(service)
		ctx := context.Background()
		go func() {
			if err := consumer.StartConsuming(ctx); err != nil {
				log.Printf("Failed to start Kafka consumer: %v", err)
			}
		}()
	}

	// Setup graceful shutdown
	defer func() {
		log.Println("Shutting down...")
		arena.CloseKafkaPublisher()
		arena.CloseKafkaConsumer()
		arena.CloseWarriorClient()
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

	// Setup health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "arena",
		})
	})

	// Setup routes
	arena.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	log.Printf("Arena HTTP service starting on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

