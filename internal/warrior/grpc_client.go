package warrior

import (
    "fmt"
    "os"
    pbWeapon "network-sec-micro/api/proto/weapon"
    pbRepair "network-sec-micro/api/proto/repair"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var weaponGrpcClient pbWeapon.WeaponServiceClient
var weaponGrpcConn *grpc.ClientConn
var repairGrpcClient pbRepair.RepairServiceClient
var repairGrpcConn *grpc.ClientConn

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

func GetWeaponClient() pbWeapon.WeaponServiceClient { return weaponGrpcClient }
func GetRepairClient() pbRepair.RepairServiceClient { return repairGrpcClient }
func CloseWeaponClient() { if weaponGrpcConn != nil { weaponGrpcConn.Close() } }
func CloseRepairClient() { if repairGrpcConn != nil { repairGrpcConn.Close() } }


