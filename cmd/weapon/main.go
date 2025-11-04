// @title Weapon Service API
// @version 1.0
// @description Weapon service handles weapon creation, purchase, and management. Warriors can buy weapons which trigger Kafka events for coin deduction.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"os"
    "net"
    "sync"

	"network-sec-micro/internal/weapon"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
    pbWeapon "network-sec-micro/api/proto/weapon"
    "google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := weapon.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize service and handler
	service := weapon.NewService()
	handler := weapon.NewHandler(service)

	// Setup graceful shutdown
	defer func() {
		log.Println("Shutting down...")
		weapon.CloseKafkaPublisher()
	}()

	// Set Gin to release mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

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
			"service": "weapon",
		})
	})

	// Setup routes
	weapon.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // Start HTTP and gRPC servers
    httpPort := os.Getenv("PORT"); if httpPort == "" { httpPort = "8081" }
    grpcPort := os.Getenv("GRPC_PORT"); if grpcPort == "" { grpcPort = "50057" }

    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        log.Printf("Weapon HTTP service starting on :%s", httpPort)
        if err := r.Run(":" + httpPort); err != nil {
            log.Fatalf("Failed to start HTTP server: %v", err)
        }
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        lis, err := net.Listen("tcp", ":"+grpcPort)
        if err != nil { log.Fatalf("gRPC listen error: %v", err) }
        s := grpc.NewServer()
        pbWeapon.RegisterWeaponServiceServer(s, weapon.NewWeaponServiceServer())
        log.Printf("Weapon gRPC service starting on :%s", grpcPort)
        if err := s.Serve(lis); err != nil { log.Fatalf("gRPC serve error: %v", err) }
    }()

    wg.Wait()
}