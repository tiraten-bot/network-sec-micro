package coin

import (
    "fmt"
    "os"

    pbWarrior "network-sec-micro/api/proto/warrior"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn

func InitWarriorClient() error {
    addr := os.Getenv("WARRIOR_GRPC_ADDR")
    if addr == "" { addr = "localhost:50052" }
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { return fmt.Errorf("failed to connect to warrior gRPC: %w", err) }
    warriorGrpcConn = conn
    warriorGrpcClient = pbWarrior.NewWarriorServiceClient(conn)
    return nil
}

func CloseWarriorClient() {
    if warriorGrpcConn != nil { warriorGrpcConn.Close() }
}

func GetWarriorByID(warriorID uint) (*pbWarrior.Warrior, error) {
    if warriorGrpcClient == nil { return nil, fmt.Errorf("warrior gRPC client not initialized") }
    req := &pbWarrior.GetWarriorByIDRequest{WarriorId: uint32(warriorID)}
    resp, err := warriorGrpcClient.GetWarriorByID(contextBackground())
    _ = req
    _ = resp
    _ = err
    return nil, fmt.Errorf("not implemented")
}

func contextBackground() interface{} { return nil }


