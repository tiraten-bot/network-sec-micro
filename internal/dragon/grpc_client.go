package dragon

import (
	"context"
	"fmt"
	"log"
	"os"

	pbRepair "network-sec-micro/api/proto/repair"
	pbWarrior "network-sec-micro/api/proto/warrior"
	pbWeapon "network-sec-micro/api/proto/weapon"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var weaponGrpcClient pbWeapon.WeaponServiceClient
var weaponGrpcConn *grpc.ClientConn
var repairGrpcClient pbRepair.RepairServiceClient
var repairGrpcConn *grpc.ClientConn

// Warrior gRPC client wrapper
type WarriorClient struct {
	conn   *grpc.ClientConn
	client pbWarrior.WarriorServiceClient
}

var warriorClient *WarriorClient

func InitWeaponClient(addr string) error {
	if addr == "" { addr = os.Getenv("WEAPON_GRPC_ADDR"); if addr == "" { addr = "localhost:50057" } }
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { return fmt.Errorf("failed to connect to weapon gRPC: %w", err) }
	weaponGrpcConn = conn
	weaponGrpcClient = pbWeapon.NewWeaponServiceClient(conn)
	return nil
}

func InitRepairClient(addr string) error {
	if addr == "" { addr = os.Getenv("REPAIR_GRPC_ADDR"); if addr == "" { addr = "localhost:50061" } }
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { return fmt.Errorf("failed to connect to repair gRPC: %w", err) }
	repairGrpcConn = conn
	repairGrpcClient = pbRepair.NewRepairServiceClient(conn)
	return nil
}

// InitWarriorClient initializes gRPC client for Warrior service
func InitWarriorClient(addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to warrior service: %w", err)
	}
	warriorClient = &WarriorClient{conn: conn, client: pbWarrior.NewWarriorServiceClient(conn)}
	log.Printf("Connected to Warrior gRPC service at %s", addr)
	return nil
}

func GetWeaponClient() pbWeapon.WeaponServiceClient { return weaponGrpcClient }
func GetRepairClient() pbRepair.RepairServiceClient { return repairGrpcClient }
func CloseWeaponClient() { if weaponGrpcConn != nil { weaponGrpcConn.Close() } }
func CloseRepairClient() { if repairGrpcConn != nil { repairGrpcConn.Close() } }

// GetWarriorClient returns the warrior client instance
func GetWarriorClient() *WarriorClient {
	if warriorClient == nil {
		addr := os.Getenv("WARRIOR_GRPC_HOST")
		if addr == "" { addr = "localhost:50052" }
		_ = InitWarriorClient(addr)
	}
	return warriorClient
}

// GetWarriorByUsername gets warrior by username
func (c *WarriorClient) GetWarriorByUsername(ctx context.Context, username string) (*pbWarrior.Warrior, error) {
	req := &pbWarrior.GetWarriorByUsernameRequest{Username: username}
	resp, err := c.client.GetWarriorByUsername(ctx, req)
	if err != nil { return nil, fmt.Errorf("failed to get warrior by username: %w", err) }
	return resp.Warrior, nil
}

// GetWarriorByID gets warrior by ID
func (c *WarriorClient) GetWarriorByID(ctx context.Context, id uint32) (*pbWarrior.Warrior, error) {
	req := &pbWarrior.GetWarriorByIDRequest{WarriorId: id}
	resp, err := c.client.GetWarriorByID(ctx, req)
	if err != nil { return nil, fmt.Errorf("failed to get warrior by ID: %w", err) }
	return resp.Warrior, nil
}

// UpdateWarriorPower updates warrior's power
func (c *WarriorClient) UpdateWarriorPower(ctx context.Context, id uint32, power int32) error {
	req := &pbWarrior.UpdateWarriorPowerRequest{WarriorId: id, TotalPower: power, WeaponCount: 0}
	_, err := c.client.UpdateWarriorPower(ctx, req)
	if err != nil { return fmt.Errorf("failed to update warrior power: %w", err) }
	return nil
}

// Close closes the gRPC connection
func (c *WarriorClient) Close() error {
	if c.conn != nil { return c.conn.Close() }
	return nil
}
