// @title Warrior Service API
// @version 1.0
// @description Warrior service handles user authentication, registration, and warrior management with role-based access control (RBAC)
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
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

	"network-sec-micro/internal/warrior"
    kafkaLib "network-sec-micro/pkg/kafka"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database first
	if err := warrior.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

    // Initialize Kafka consumer for achievements
    brokers := getEnvSlice("KAFKA_BROKERS", "localhost:9092")
    consumer, err := kafkaLib.NewConsumer(
        brokers,
        "warrior-service-group",
        []string{kafkaLib.TopicDragonDeath, kafkaLib.TopicEnemyDestroyed},
        warrior.ProcessKafkaMessage,
    )
    if err != nil {
        log.Fatalf("Failed to create Kafka consumer: %v", err)
    }
    defer consumer.Close()
    if err := consumer.Start(); err != nil {
        log.Fatalf("Failed to start Kafka consumer: %v", err)
    }

	// Initialize dependencies manually (Wire has dependency issues with puddle/v2)
	service := warrior.NewService()
	handler := warrior.NewHandler(service)

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
			"service": "warrior",
		})
	})

	// Setup routes with manually injected handler
	warrior.SetupRoutes(r, handler)

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Warrior service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnvSlice(key, defaultValue string) []string {
    value := os.Getenv(key)
    if value == "" {
        return []string{defaultValue}
    }
    var out []string
    for _, s := range splitAndTrim(value, ",") {
        if s != "" {
            out = append(out, s)
        }
    }
    if len(out) == 0 {
        return []string{defaultValue}
    }
    return out
}

func splitAndTrim(s, sep string) []string {
    parts := []string{}
    start := 0
    for i := 0; i <= len(s); i++ {
        if i == len(s) || string(s[i]) == sep {
            part := s[start:i]
            // trim spaces
            for len(part) > 0 && (part[0] == ' ' || part[0] == '\t') {
                part = part[1:]
            }
            for len(part) > 0 && (part[len(part)-1] == ' ' || part[len(part)-1] == '\t') {
                part = part[:len(part)-1]
            }
            parts = append(parts, part)
            start = i + 1
        }
    }
    return parts
}
