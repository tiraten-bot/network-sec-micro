// @title Battle Service API
// @version 1.0
// @description Battle service handles turn-based combat between warriors and enemies/dragons. Manages battle history, statistics, and rewards.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8085
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/battle"
	"network-sec-micro/internal/battle"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := battle.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

    // Optionally initialize PostgreSQL (gradual migration)
    if err := battle.InitPostgres(); err != nil {
        log.Printf("Warning: Battle Postgres init failed: %v", err)
    }

    // Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}

	if err := battle.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Initialize Coin gRPC client
	coinAddr := os.Getenv("COIN_GRPC_ADDR")
	if coinAddr == "" {
		coinAddr = "localhost:50051"
	}

	if err := battle.InitCoinClient(coinAddr); err != nil {
		log.Fatalf("Failed to connect to Coin gRPC: %v", err)
	}

	// Initialize BattleSpell gRPC client
	battlespellAddr := os.Getenv("BATTLESPELL_GRPC_ADDR")
	if battlespellAddr == "" {
		battlespellAddr = "localhost:50054"
	}

	if err := battle.InitBattlespellClient(battlespellAddr); err != nil {
		log.Fatalf("Failed to connect to BattleSpell gRPC: %v", err)
	}

	// Initialize Weapon gRPC client (optional)
	weaponAddr := os.Getenv("WEAPON_GRPC_ADDR")
	if weaponAddr == "" {
		weaponAddr = "localhost:50057"
	}
	if err := battle.InitWeaponClient(weaponAddr); err != nil {
		log.Printf("Warning: Failed to connect to Weapon gRPC: %v", err)
	}

	// Initialize Redis client for battle logs
	if err := battle.InitRedisClient(); err != nil {
		log.Printf("Warning: Failed to connect to Redis (battle logs will not be available): %v", err)
		// Don't fail startup if Redis is not available
	}

	// Initialize Heal gRPC client (optional, for healing state checks)
	healAddr := os.Getenv("HEAL_GRPC_ADDR")
	if healAddr == "" {
		healAddr = "localhost:50058"
	}
	if err := battle.InitHealClient(healAddr); err != nil {
		log.Printf("Warning: Failed to connect to Heal gRPC: %v", err)
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
		battle.CloseKafkaPublisher()
		battle.CloseWarriorClient()
		battle.CloseCoinClient()
		battle.CloseBattlespellClient()
		battle.CloseRedisClient()
		battle.CloseWeaponClient()
		battle.CloseHealClient()
	}()

	// Start gRPC server in a goroutine
	go func() {
		grpcPort := os.Getenv("GRPC_PORT")
		if grpcPort == "" {
			grpcPort = "50053"
		}

		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterBattleServiceServer(s, grpcServer)

		log.Printf("Battle gRPC service starting on port %s", grpcPort)
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

	// Setup health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "battle",
		})
	})

	// Setup routes
	battle.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("Battle HTTP service starting on port %s", port)
	
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

