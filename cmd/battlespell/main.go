package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/battlespell"
	"network-sec-micro/internal/battlespell"
	docs "network-sec-micro/cmd/battlespell/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
    _ = docs.SwaggerInfo // ensure docs package is linked
	// Initialize database
	if err := battlespell.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Battle gRPC client
	battleAddr := os.Getenv("BATTLE_GRPC_ADDR")
	if battleAddr == "" {
		battleAddr = "localhost:50053"
	}

	if err := battlespell.InitBattleClient(battleAddr); err != nil {
		log.Fatalf("Failed to connect to Battle gRPC: %v", err)
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

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	defer func() {
		log.Println("Shutting down...")
		battlespell.CloseBattleClient()
	}()

	// Start gRPC server in a goroutine
	go func() {
		grpcPort := os.Getenv("GRPC_PORT")
		if grpcPort == "" {
			grpcPort = "50054"
		}

		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterBattleSpellServiceServer(s, grpcServer)

		log.Printf("BattleSpell gRPC service starting on port %s", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
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

	// Setup routes (includes health check and metrics)
	battlespell.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("BattleSpell HTTP service starting on port %s", port)

	// Start HTTP server in a goroutine
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutdown signal received, gracefully shutting down...")
}

