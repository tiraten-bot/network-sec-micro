package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/arena"
	"network-sec-micro/internal/arena"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
    // Initialize database (Mongo, existing)
    if err := arena.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

    // Initialize Redis client (for Redis-backed repo)
    if err := arena.InitRedisClient(); err != nil {
        log.Printf("Warning: Arena Redis init failed: %v", err)
    }

    // Optionally initialize PostgreSQL (gradual migration)
    if err := arena.InitPostgres(); err != nil {
        log.Printf("Warning: Arena Postgres init failed: %v", err)
    }

    // Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}

    if err := arena.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

    // Initialize ArenaSpell gRPC client
    arenaspellAddr := os.Getenv("ARENASPELL_GRPC_ADDR")
    if arenaspellAddr == "" {
        arenaspellAddr = "localhost:50056"
    }
    if err := arena.InitArenaSpellClient(arenaspellAddr); err != nil {
        log.Fatalf("Failed to connect to ArenaSpell gRPC: %v", err)
    }

	// Set Gin to release mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize service, handler, and gRPC server using Wire
	service, handler, grpcServer, err := InitializeApp()
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
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	defer func() {
		log.Println("Shutting down...")
		arena.CloseKafkaPublisher()
		arena.CloseKafkaConsumer()
		arena.CloseWarriorClient()
        arena.CloseArenaSpellClient()
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
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8087"
	}

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50055"
	}

	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterArenaServiceServer(grpcSrv, grpcServer)

	// Start gRPC server in goroutine
	go func() {
		log.Printf("Arena gRPC service starting on port %s", grpcPort)
		if err := grpcSrv.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Arena HTTP service starting on port %s", httpPort)
		if err := r.Run(":" + httpPort); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-shutdown
	log.Println("Shutting down servers...")

	// Graceful shutdown
	grpcSrv.GracefulStop()
	log.Println("Arena service stopped")
}

