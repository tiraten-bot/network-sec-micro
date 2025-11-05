package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pb "network-sec-micro/api/proto/coin"
	"network-sec-micro/internal/coin"
	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"
	kafkaLib "network-sec-micro/pkg/kafka"

	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := coin.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Kafka consumer
	kafkaBrokers := getEnvSlice("KAFKA_BROKERS", "localhost:9092")
	consumer, err := kafkaLib.NewConsumer(
		kafkaBrokers,
		"coin-service-group",
		[]string{kafkaLib.TopicWeaponPurchase, kafkaLib.TopicArenaMatchCompleted, kafkaLib.TopicBattleWagerResolved},
		coin.ProcessKafkaMessage,
	)
	// Init Warrior gRPC client for event-driven coin awards
	if err := coin.InitWarriorClient(); err != nil {
		log.Printf("Warning: coin couldn't init warrior client: %v", err)
	}
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("Starting Kafka consumer...")
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Initialize service and gRPC server (Wire will be added later)
	service := coin.NewService()
	grpcServer := coin.NewCoinServiceServer(service)
	
	// TODO: Wire integration when wire issue is resolved
	// service, grpcServer, err := InitializeCoinApp()
	// if err != nil {
	// 	log.Fatalf("Failed to initialize app: %v", err)
	// }

	// Start gRPC server
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start metrics server
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "8091"
	}
	healthHandler := health.NewHandler(&health.DatabaseChecker{DB: coin.DB, DBName: "mysql"})
	go func() {
		if err := metrics.StartMetricsServer(metricsPort, healthHandler); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	s := grpc.NewServer()
	
	// Register coin service
	pb.RegisterCoinServiceServer(s, grpcServer)

	log.Printf("Coin gRPC service starting on port %s", port)
	log.Printf("Coin metrics server starting on port %s", metricsPort)

	// Start gRPC server in goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutdown signal received, gracefully shutting down...")
	s.GracefulStop()
	log.Println("Coin service stopped")
}

func getEnvSlice(key, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{defaultValue}
	}
	return strings.Split(value, ",")
}
