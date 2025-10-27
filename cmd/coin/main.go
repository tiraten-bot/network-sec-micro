package main

import (
	"log"
	"net"
	"os"
	"strings"

	pb "network-sec-micro/api/proto/coin"
	"network-sec-micro/internal/coin"
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
		[]string{kafkaLib.TopicWeaponPurchase},
		coin.ProcessKafkaMessage,
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("Starting Kafka consumer...")
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Initialize service
	service := coin.NewService()
	
	// Create gRPC server
	grpcServer := coin.NewCoinServiceServer(service)

	// Start gRPC server
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	
	// Register coin service
	pb.RegisterCoinServiceServer(s, grpcServer)

	log.Printf("Coin gRPC service starting on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnvSlice(key, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{defaultValue}
	}
	return strings.Split(value, ",")
}
