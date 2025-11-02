package battlespell

import (
	"context"
	"fmt"
	"log"
	"os"

	pbBattle "network-sec-micro/api/proto/battle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var battleGrpcClient pbBattle.BattleServiceClient
var battleGrpcConn *grpc.ClientConn

// InitBattleClient initializes the gRPC client connection to battle service
func InitBattleClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("BATTLE_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50053"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to battle gRPC: %w", err)
	}

	battleGrpcClient = pbBattle.NewBattleServiceClient(conn)
	battleGrpcConn = conn

	log.Printf("Connected to Battle gRPC service at %s", addr)
	return nil
}

// CloseBattleClient closes the gRPC connection
func CloseBattleClient() {
	if battleGrpcConn != nil {
		battleGrpcConn.Close()
	}
}

// GetBattleClient returns the battle gRPC client
func GetBattleClient() pbBattle.BattleServiceClient {
	return battleGrpcClient
}

// UpdateParticipantStats updates participant stats via battle service gRPC
func UpdateParticipantStats(ctx context.Context, battleID, participantID string, hp, maxHP, attackPower, defense int32, isAlive bool) error {
	if battleGrpcClient == nil {
		return fmt.Errorf("battle gRPC client not initialized")
	}

	req := &pbBattle.UpdateParticipantStatsRequest{
		BattleId:      battleID,
		ParticipantId: participantID,
		Hp:            hp,
		MaxHp:         maxHP,
		AttackPower:   attackPower,
		Defense:       defense,
		IsAlive:       isAlive,
	}

	_, err := battleGrpcClient.UpdateParticipantStats(ctx, req)
	return err
}

// GetBattleParticipants gets battle participants via battle service gRPC
func GetBattleParticipants(ctx context.Context, battleID, side string) ([]*pbBattle.BattleParticipant, error) {
	if battleGrpcClient == nil {
		return nil, fmt.Errorf("battle gRPC client not initialized")
	}

	req := &pbBattle.GetBattleParticipantsRequest{
		BattleId: battleID,
		Side:      side,
	}

	resp, err := battleGrpcClient.GetBattleParticipants(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Participants, nil
}

// GetBattleByID gets battle by ID via battle service gRPC
func GetBattleByID(ctx context.Context, battleID string) (*pbBattle.Battle, error) {
	if battleGrpcClient == nil {
		return nil, fmt.Errorf("battle gRPC client not initialized")
	}

	req := &pbBattle.GetBattleByIDRequest{
		BattleId: battleID,
	}

	resp, err := battleGrpcClient.GetBattleByID(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Battle, nil
}

