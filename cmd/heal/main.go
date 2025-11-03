package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/heal"
	"network-sec-micro/internal/heal"

	"google.golang.org/grpc"
)

func main() {
	// Initialize PostgreSQL (optional, for healing records storage)
	if err := heal.InitPostgres(); err != nil {
		log.Printf("Warning: Heal Postgres init failed: %v", err)
	}

	// Initialize Redis client
	if err := heal.InitRedisClient(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	}

	// Initialize Warrior gRPC client
	warriorAddr := os.Getenv("WARRIOR_GRPC_ADDR")
	if warriorAddr == "" {
		warriorAddr = "localhost:50052"
	}
	if err := heal.InitWarriorClient(warriorAddr); err != nil {
		log.Fatalf("Failed to connect to Warrior gRPC: %v", err)
	}

	// Initialize Coin gRPC client
	coinAddr := os.Getenv("COIN_GRPC_ADDR")
	if coinAddr == "" {
		coinAddr = "localhost:50051"
	}
	if err := heal.InitCoinClient(coinAddr); err != nil {
		log.Fatalf("Failed to connect to Coin gRPC: %v", err)
	}

	// Initialize service and gRPC server using Wire
	_, grpcServer, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	defer func() {
		log.Println("Shutting down...")
		heal.CloseWarriorClient()
		heal.CloseCoinClient()
		heal.CloseRedisClient()
	}()

	// Start Kafka consumer in background
	go func() {
		if err := heal.StartKafkaConsumer(); err != nil {
			log.Printf("Warning: Failed to start Kafka consumer: %v", err)
		}
	}()

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50058"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHealServiceServer(s, grpcServer)

	log.Printf("Heal gRPC service starting on port %s", grpcPort)

	// Start gRPC server in goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutdown signal received, gracefully shutting down...")
	s.GracefulStop()
	log.Println("Heal service stopped")
}

