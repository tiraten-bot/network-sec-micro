package battle

import (
	"context"
	"fmt"
	"log"

	pb "network-sec-micro/api/proto/warrior"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcClient pb.WarriorServiceClient
var grpcConn *grpc.ClientConn

// InitWarriorClient initializes the gRPC client connection to warrior service
func InitWarriorClient(addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to warrior gRPC: %w", err)
	}

	grpcClient = pb.NewWarriorServiceClient(conn)
	grpcConn = conn

	log.Printf("Connected to Warrior gRPC service at %s", addr)
	return nil
}

// CloseWarriorClient closes the gRPC connection
func CloseWarriorClient() {
	if grpcConn != nil {
		grpcConn.Close()
	}
}

// GetWarriorByUsername gets warrior info via gRPC
func GetWarriorByUsername(ctx context.Context, username string) (*pb.Warrior, error) {
	if grpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pb.GetWarriorByUsernameRequest{
		Username: username,
	}

	resp, err := grpcClient.GetWarriorByUsername(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	return resp.Warrior, nil
}

// GetWarriorByID gets warrior info by ID via gRPC
func GetWarriorByID(ctx context.Context, warriorID uint) (*pb.Warrior, error) {
	if grpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pb.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	}

	resp, err := grpcClient.GetWarriorByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	return resp.Warrior, nil
}

