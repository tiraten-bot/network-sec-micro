package enemy

import (
	"context"
	"log"

	pbWarrior "network-sec-micro/api/proto/warrior"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorClient pbWarrior.WarriorServiceClient

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

// GetWarriorBalance gets warrior balance (for goblin attacks)
func GetWarriorBalance(ctx context.Context, warriorID uint) (int32, error) {
	// Note: We'll use warrior proto for getting balance
	// Since warrior proto has coin_balance field
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

