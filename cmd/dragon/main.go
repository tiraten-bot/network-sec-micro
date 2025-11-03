// @title Dragon Service API
// @version 1.0
// @description Dragon service handles dragon creation, attack management, and death events. Dragons can be attacked by warriors, and death events are published to Kafka.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8084
// @BasePath /api/v1
// @schemes http https

package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"network-sec-micro/internal/dragon"
	pb "network-sec-micro/api/proto/dragon"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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
	grpcServer := dragon.NewDragonServiceServer(service)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50059"
	}
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterDragonServiceServer(grpcSrv, grpcServer)

	go func() {
		log.Printf("Dragon gRPC server starting on port %s", grpcPort)
		if err := grpcSrv.Serve(grpcLis); err != nil {
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

	// Setup routes
	dragon.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Dragon HTTP service starting on port %s", port)
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down Dragon service...")
	grpcSrv.GracefulStop()
	dragon.CloseKafkaPublisher()
	log.Println("Dragon service stopped")
}
