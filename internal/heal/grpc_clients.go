package heal

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
	pbCoin "network-sec-micro/api/proto/coin"
	pbDragon "network-sec-micro/api/proto/dragon"
	pbEnemy "network-sec-micro/api/proto/enemy"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn

var coinGrpcClient pbCoin.CoinServiceClient
var coinGrpcConn *grpc.ClientConn

var dragonGrpcClient pbDragon.DragonServiceClient
var dragonGrpcConn *grpc.ClientConn

var enemyGrpcClient pbEnemy.EnemyServiceClient
var enemyGrpcConn *grpc.ClientConn

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

// CloseWarriorClient closes the gRPC connection
func CloseWarriorClient() {
	if warriorGrpcConn != nil {
		warriorGrpcConn.Close()
	}
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

// CloseCoinClient closes the coin gRPC connection
func CloseCoinClient() {
	if coinGrpcConn != nil {
		coinGrpcConn.Close()
	}
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

// UpdateWarriorHP updates warrior's HP via gRPC
func UpdateWarriorHP(ctx context.Context, warriorID uint, newHP int32) error {
	if warriorGrpcClient == nil {
		return fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.UpdateWarriorHPRequest{
		WarriorId: uint32(warriorID),
		NewHp:     newHP,
	}

	resp, err := warriorGrpcClient.UpdateWarriorHP(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update warrior HP: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update warrior HP: %s", resp.Message)
	}

	log.Printf("Updated warrior %d HP: %d -> %d", warriorID, resp.OldHp, resp.NewHp)
	return nil
}

// SetWarriorHealingState sets warrior's healing state (is_healing, healing_until)
func SetWarriorHealingState(ctx context.Context, warriorID uint, isHealing bool, healingUntil *time.Time) error {
	if warriorGrpcClient == nil {
		return fmt.Errorf("warrior gRPC client not initialized")
	}

	var healingUntilSeconds int64
	if healingUntil != nil {
		healingUntilSeconds = healingUntil.Unix()
	}

	req := &pbWarrior.UpdateWarriorHealingStateRequest{
		WarriorId:          uint32(warriorID),
		IsHealing:          isHealing,
		HealingUntilSeconds: healingUntilSeconds,
	}

	resp, err := warriorGrpcClient.UpdateWarriorHealingState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update warrior healing state: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update warrior healing state: %s", resp.Message)
	}

	if isHealing && healingUntil != nil {
		log.Printf("Warrior %d is now healing until %v", warriorID, healingUntil)
	} else {
		log.Printf("Warrior %d healing completed", warriorID)
	}
	return nil
}

// CheckWarriorHealingState checks if warrior is currently healing
func CheckWarriorHealingState(ctx context.Context, warriorID uint) (bool, *time.Time, error) {
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return false, nil, err
	}

	if !warrior.IsHealing || warrior.HealingUntilSeconds == 0 {
		return false, nil, nil
	}

	healingUntil := time.Unix(warrior.HealingUntilSeconds, 0)
	now := time.Now()
	if now.After(healingUntil) {
		// Healing time passed, clear state
		_ = SetWarriorHealingState(ctx, warriorID, false, nil)
		return false, nil, nil
	}

	return true, &healingUntil, nil
}

