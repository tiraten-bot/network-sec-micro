package battle

import (
	"context"
	"fmt"
	"log"
	"os"

	pbWarrior "network-sec-micro/api/proto/warrior"
	pbCoin "network-sec-micro/api/proto/coin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn

var coinGrpcClient pbCoin.CoinServiceClient
var coinGrpcConn *grpc.ClientConn

// InitWarriorClient initializes the gRPC client connection to warrior service
func InitWarriorClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("WARRIOR_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50052"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to warrior gRPC: %w", err)
	}

	warriorGrpcClient = pbWarrior.NewWarriorServiceClient(conn)
	warriorGrpcConn = conn

	log.Printf("Connected to Warrior gRPC service at %s", addr)
	return nil
}

// InitCoinClient initializes the gRPC client connection to coin service
func InitCoinClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("COIN_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50051"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to coin gRPC: %w", err)
	}

	coinGrpcClient = pbCoin.NewCoinServiceClient(conn)
	coinGrpcConn = conn

	log.Printf("Connected to Coin gRPC service at %s", addr)
	return nil
}

// CloseWarriorClient closes the gRPC connection
func CloseWarriorClient() {
	if warriorGrpcConn != nil {
		warriorGrpcConn.Close()
	}
}

// CloseCoinClient closes the coin gRPC connection
func CloseCoinClient() {
	if coinGrpcConn != nil {
		coinGrpcConn.Close()
	}
}

// GetWarriorByUsername gets warrior info via gRPC
func GetWarriorByUsername(ctx context.Context, username string) (*pbWarrior.Warrior, error) {
	if warriorGrpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.GetWarriorByUsernameRequest{
		Username: username,
	}

	resp, err := warriorGrpcClient.GetWarriorByUsername(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	return resp.Warrior, nil
}

// GetWarriorByID gets warrior info by ID via gRPC
func GetWarriorByID(ctx context.Context, warriorID uint) (*pbWarrior.Warrior, error) {
	if warriorGrpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	}

	resp, err := warriorGrpcClient.GetWarriorByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	return resp.Warrior, nil
}

// AddCoins adds coins to warrior's balance via gRPC
func AddCoins(ctx context.Context, warriorID uint, amount int64, reason string) error {
	if coinGrpcClient == nil {
		return fmt.Errorf("coin gRPC client not initialized")
	}

	req := &pbCoin.AddCoinsRequest{
		WarriorId: uint32(warriorID),
		Amount:    amount,
		Reason:    reason,
	}

	resp, err := coinGrpcClient.AddCoins(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to add coins: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to add coins: %s", resp.Message)
	}

	log.Printf("Added %d coins to warrior %d. Balance: %d -> %d", amount, warriorID, resp.BalanceBefore, resp.BalanceAfter)
	return nil
}

// DeductCoins deducts coins from warrior's balance via gRPC
func DeductCoins(ctx context.Context, warriorID uint, amount int64, reason string) error {
	if coinGrpcClient == nil {
		return fmt.Errorf("coin gRPC client not initialized")
	}

	req := &pbCoin.DeductCoinsRequest{
		WarriorId: uint32(warriorID),
		Amount:    amount,
		Reason:    reason,
	}

	resp, err := coinGrpcClient.DeductCoins(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to deduct coins: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to deduct coins: %s", resp.Message)
	}

	log.Printf("Deducted %d coins from warrior %d. Balance: %d -> %d", amount, warriorID, resp.BalanceBefore, resp.BalanceAfter)
	return nil
}

