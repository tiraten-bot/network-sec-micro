package dragon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"network-sec-micro/api/proto/warrior"
	pb "network-sec-micro/api/proto/warrior"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// WarriorClient handles gRPC communication with Warrior service
type WarriorClient struct {
	conn   *grpc.ClientConn
	client pb.WarriorServiceClient
}

var warriorClient *WarriorClient

// InitWarriorClient initializes gRPC client for Warrior service
func InitWarriorClient(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to warrior service: %w", err)
	}

	warriorClient = &WarriorClient{
		conn:   conn,
		client: pb.NewWarriorServiceClient(conn),
	}

	log.Printf("Connected to Warrior gRPC service at %s", addr)
	return nil
}

// GetWarriorClient returns the warrior client instance
func GetWarriorClient() *WarriorClient {
	if warriorClient == nil {
		addr := os.Getenv("WARRIOR_GRPC_HOST")
		if addr == "" {
			addr = "localhost:50052"
		}
		InitWarriorClient(addr)
	}
	return warriorClient
}

// GetWarriorByUsername gets warrior by username
func (c *WarriorClient) GetWarriorByUsername(ctx context.Context, username string) (*warrior.Warrior, error) {
	req := &pb.GetWarriorByUsernameRequest{
		Username: username,
	}

	resp, err := c.client.GetWarriorByUsername(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior by username: %w", err)
	}

	return resp.Warrior, nil
}

// GetWarriorByID gets warrior by ID
func (c *WarriorClient) GetWarriorByID(ctx context.Context, id uint32) (*warrior.Warrior, error) {
	req := &pb.GetWarriorByIDRequest{
		WarriorId: id,
	}

	resp, err := c.client.GetWarriorByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior by ID: %w", err)
	}

	return resp.Warrior, nil
}

// UpdateWarriorPower updates warrior's power
func (c *WarriorClient) UpdateWarriorPower(ctx context.Context, id uint32, power int32) error {
	req := &pb.UpdateWarriorPowerRequest{
		WarriorId:   id,
		TotalPower:  power,
		WeaponCount: 0, // Not used in this context
	}

	_, err := c.client.UpdateWarriorPower(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update warrior power: %w", err)
	}

	return nil
}

// Close closes the gRPC connection
func (c *WarriorClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
