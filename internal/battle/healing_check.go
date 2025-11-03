package battle

import (
	"context"
	"fmt"
	"time"

	pbHeal "network-sec-micro/api/proto/heal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

var healGrpcClient pbHeal.HealServiceClient
var healGrpcConn *grpc.ClientConn

// InitHealClient initializes the gRPC client connection to heal service
func InitHealClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("HEAL_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50058"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to heal gRPC: %w", err)
	}

	healGrpcClient = pbHeal.NewHealServiceClient(conn)
	healGrpcConn = conn
	return nil
}

// CloseHealClient closes the heal gRPC connection
func CloseHealClient() {
	if healGrpcConn != nil {
		healGrpcConn.Close()
	}
}

// CheckWarriorCanBattle checks if warrior is currently healing (can't battle while healing)
func CheckWarriorCanBattle(ctx context.Context, warriorID uint) error {
	if healGrpcClient == nil {
		// If heal service is not available, skip check
		return nil
	}

	// Get warrior info to check healing state
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return fmt.Errorf("failed to get warrior: %w", err)
	}

	if warrior.IsHealing && warrior.HealingUntilSeconds > 0 {
		healingUntil := time.Unix(warrior.HealingUntilSeconds, 0)
		if time.Now().Before(healingUntil) {
			remaining := time.Until(healingUntil).Seconds()
			return fmt.Errorf("warrior is currently healing. Cannot start battle. Remaining time: %.0f seconds", remaining)
		}
	}

	return nil
}

