package enemy

import (
	"context"
	"time"

	pb "network-sec-micro/api/proto/enemy"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EnemyServiceServer implements the EnemyService gRPC interface
type EnemyServiceServer struct {
	pb.UnimplementedEnemyServiceServer
	Service *Service
}

// NewEnemyServiceServer creates a new enemy gRPC server
func NewEnemyServiceServer(service *Service) *EnemyServiceServer {
	return &EnemyServiceServer{
		Service: service,
	}
}

// GetEnemyByID retrieves an enemy by ID
func (s *EnemyServiceServer) GetEnemyByID(ctx context.Context, req *pb.GetEnemyByIDRequest) (*pb.GetEnemyByIDResponse, error) {
	enemyID, err := primitive.ObjectIDFromHex(req.EnemyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid enemy ID: %v", err)
	}

	var enemy Enemy
	if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "enemy not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get enemy: %v", err)
	}

	// Set MaxHealth if not set (use Health as default)
	maxHealth := enemy.MaxHealth
	if maxHealth == 0 {
		maxHealth = enemy.Health
	}

	healingUntilSeconds := int64(0)
	if enemy.HealingUntil != nil {
		healingUntilSeconds = enemy.HealingUntil.Unix()
	}

	return &pb.GetEnemyByIDResponse{
		Enemy: &pb.Enemy{
			Id:                 enemy.ID.Hex(),
			Name:               enemy.Name,
			Type:               string(enemy.Type),
		Level:              int32(enemy.Level),
		Health:             int32(enemy.Health),
		MaxHealth:          int32(maxHealth),
		AttackPower:        int32(enemy.AttackPower),
		CoinBalance:        enemy.CoinBalance,
		CreatedBy:          enemy.CreatedBy,
		IsHealing:          enemy.IsHealing,
		HealingUntilSeconds: healingUntilSeconds,
		},
		Success: true,
		Message: "enemy retrieved successfully",
	}, nil
}

// UpdateEnemyHP updates an enemy's HP
func (s *EnemyServiceServer) UpdateEnemyHP(ctx context.Context, req *pb.UpdateEnemyHPRequest) (*pb.UpdateEnemyHPResponse, error) {
	enemyID, err := primitive.ObjectIDFromHex(req.EnemyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid enemy ID: %v", err)
	}

	var enemy Enemy
	if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "enemy not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get enemy: %v", err)
	}

	// Set MaxHealth if not set
	maxHealth := enemy.MaxHealth
	if maxHealth == 0 {
		maxHealth = enemy.Health
	}

	oldHP := enemy.Health
	newHP := int(req.NewHp)
	if newHP > maxHealth {
		newHP = maxHealth
	}
	if newHP < 0 {
		newHP = 0
	}

	updateData := bson.M{
		"health":     newHP,
		"updated_at": time.Now(),
	}

	// Update MaxHealth if it wasn't set before
	if enemy.MaxHealth == 0 {
		updateData["max_health"] = maxHealth
	}

	_, err = EnemyColl.UpdateOne(ctx, bson.M{"_id": enemyID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update enemy HP: %v", err)
	}

	return &pb.UpdateEnemyHPResponse{
		Success:   true,
		Message:   "enemy HP updated successfully",
		CurrentHp: int32(newHP),
	}, nil
}

// UpdateEnemyHealingState updates enemy's healing state
func (s *EnemyServiceServer) UpdateEnemyHealingState(ctx context.Context, req *pb.UpdateEnemyHealingStateRequest) (*pb.UpdateEnemyHealingStateResponse, error) {
	enemyID, err := primitive.ObjectIDFromHex(req.EnemyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid enemy ID: %v", err)
	}

	updateData := bson.M{
		"is_healing": req.IsHealing,
		"updated_at": time.Now(),
	}

	if req.HealingUntilSeconds > 0 {
		healingUntil := time.Unix(req.HealingUntilSeconds, 0)
		updateData["healing_until"] = healingUntil
	} else {
		updateData["healing_until"] = nil
	}

	_, err = EnemyColl.UpdateOne(ctx, bson.M{"_id": enemyID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update healing state: %v", err)
	}

	return &pb.UpdateEnemyHealingStateResponse{
		Success: true,
		Message: "healing state updated successfully",
	}, nil
}

// CheckEnemyCanBattle checks if an enemy can participate in battles
func (s *EnemyServiceServer) CheckEnemyCanBattle(ctx context.Context, req *pb.CheckEnemyCanBattleRequest) (*pb.CheckEnemyCanBattleResponse, error) {
	enemyID, err := primitive.ObjectIDFromHex(req.EnemyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid enemy ID: %v", err)
	}

	var enemy Enemy
	if err := EnemyColl.FindOne(ctx, bson.M{"_id": enemyID}).Decode(&enemy); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "enemy not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get enemy: %v", err)
	}

	if enemy.IsHealing && enemy.HealingUntil != nil {
		now := time.Now()
		if now.Before(*enemy.HealingUntil) {
			return &pb.CheckEnemyCanBattleResponse{
				CanBattle:           false,
				Reason:              "enemy is currently healing",
				HealingUntilSeconds: enemy.HealingUntil.Unix(),
			}, nil
		}
		// Healing expired, clear state
		_ = s.UpdateEnemyHealingState(ctx, &pb.UpdateEnemyHealingStateRequest{
			EnemyId:            req.EnemyId,
			IsHealing:         false,
			HealingUntilSeconds: 0,
		})
	}

	return &pb.CheckEnemyCanBattleResponse{
		CanBattle:           true,
		Reason:              "",
		HealingUntilSeconds: 0,
	}, nil
}

