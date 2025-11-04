package enemy

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
var warriorClient pbWarrior.WarriorServiceClient

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

// InitWarriorClient initializes gRPC client for warrior service
func InitWarriorClient(warriorAddr string) error {
	conn, err := grpc.Dial(warriorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	warriorClient = pbWarrior.NewWarriorServiceClient(conn)
	log.Println("Connected to Warrior Service gRPC")
	return nil
}

func GetWeaponClient() pbWeapon.WeaponServiceClient { return weaponGrpcClient }
func GetRepairClient() pbRepair.RepairServiceClient { return repairGrpcClient }
func CloseWeaponClient() { if weaponGrpcConn != nil { weaponGrpcConn.Close() } }
func CloseRepairClient() { if repairGrpcConn != nil { repairGrpcConn.Close() } }

// GetWarriorBalance gets warrior balance (for goblin attacks)
func GetWarriorBalance(ctx context.Context, warriorID uint) (int32, error) {
	resp, err := warriorClient.GetWarriorByID(ctx, &pbWarrior.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	})
	if err != nil {
		return 0, err
	}
	return resp.Warrior.CoinBalance, nil
}

// GetWarriorWeaponCount gets warrior weapon count (for pirate attacks)
func GetWarriorWeaponCount(ctx context.Context, warriorID uint) (int32, error) {
	resp, err := warriorClient.GetWarriorByID(ctx, &pbWarrior.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	})
	if err != nil {
		return 0, err
	}
	return resp.Warrior.WeaponCount, nil
}

// GetWarriorByUsername gets warrior by username
func GetWarriorByUsername(ctx context.Context, username string) (*pbWarrior.Warrior, error) {
	resp, err := warriorClient.GetWarriorByUsername(ctx, &pbWarrior.GetWarriorByUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	return resp.Warrior, nil
}

