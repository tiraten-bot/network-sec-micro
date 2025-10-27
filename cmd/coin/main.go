package main

import (
	"log"
	"net"
	"os"

	pb "network-sec-micro/api/proto/coin"
	"network-sec-micro/internal/coin"

	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	if err := coin.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
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

