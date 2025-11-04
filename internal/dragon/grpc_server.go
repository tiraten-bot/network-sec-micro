package dragon

import (
	"context"
	"time"

	pb "network-sec-micro/api/proto/dragon"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DragonServiceServer implements the DragonService gRPC interface
type DragonServiceServer struct {
	pb.UnimplementedDragonServiceServer
	Service *Service
}

// NewDragonServiceServer creates a new dragon gRPC server
func NewDragonServiceServer(service *Service) *DragonServiceServer {
	return &DragonServiceServer{
		Service: service,
	}
}

// GetDragonByID retrieves a dragon by ID
func (s *DragonServiceServer) GetDragonByID(ctx context.Context, req *pb.GetDragonByIDRequest) (*pb.GetDragonByIDResponse, error) {
	dragonID, err := primitive.ObjectIDFromHex(req.DragonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dragon ID: %v", err)
	}

	var dragon Dragon
	if err := DragonColl.FindOne(ctx, bson.M{"_id": dragonID}).Decode(&dragon); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "dragon not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get dragon: %v", err)
	}

	healingUntilSeconds := int64(0)
	if dragon.HealingUntil != nil {
		healingUntilSeconds = dragon.HealingUntil.Unix()
	}

	return &pb.GetDragonByIDResponse{
		Dragon: &pb.Dragon{
			Id:                 dragon.ID.Hex(),
			Name:               dragon.Name,
			Type:               string(dragon.Type),
			Level:              int32(dragon.Level),
			Health:             int32(dragon.Health),
			MaxHealth:          int32(dragon.MaxHealth),
			AttackPower:        int32(dragon.AttackPower),
			Defense:            int32(dragon.Defense),
			CreatedBy:          dragon.CreatedBy,
			IsAlive:            dragon.IsAlive,
			IsHealing:          dragon.IsHealing,
			HealingUntilSeconds: healingUntilSeconds,
			RevivalCount:       int32(dragon.RevivalCount),
		},
		Success: true,
		Message: "dragon retrieved successfully",
	}, nil
}

// UpdateDragonHP updates a dragon's HP
func (s *DragonServiceServer) UpdateDragonHP(ctx context.Context, req *pb.UpdateDragonHPRequest) (*pb.UpdateDragonHPResponse, error) {
	dragonID, err := primitive.ObjectIDFromHex(req.DragonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dragon ID: %v", err)
	}

	var dragon Dragon
	if err := DragonColl.FindOne(ctx, bson.M{"_id": dragonID}).Decode(&dragon); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "dragon not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get dragon: %v", err)
	}

	newHP := int(req.NewHp)
	if newHP > dragon.MaxHealth {
		newHP = dragon.MaxHealth
	}
	if newHP < 0 {
		newHP = 0
	}

	updateData := bson.M{
		"health":     newHP,
		"updated_at": time.Now(),
	}

	if newHP == 0 {
		updateData["is_alive"] = false
	} else if !dragon.IsAlive && newHP > 0 {
		updateData["is_alive"] = true
	}

	_, err = DragonColl.UpdateOne(ctx, bson.M{"_id": dragonID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update dragon HP: %v", err)
	}

	return &pb.UpdateDragonHPResponse{
		Success:   true,
		Message:   "dragon HP updated successfully",
		CurrentHp: int32(newHP),
	}, nil
}

// UpdateDragonHealingState updates dragon's healing state
func (s *DragonServiceServer) UpdateDragonHealingState(ctx context.Context, req *pb.UpdateDragonHealingStateRequest) (*pb.UpdateDragonHealingStateResponse, error) {
	dragonID, err := primitive.ObjectIDFromHex(req.DragonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dragon ID: %v", err)
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

	_, err = DragonColl.UpdateOne(ctx, bson.M{"_id": dragonID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update healing state: %v", err)
	}

	return &pb.UpdateDragonHealingStateResponse{
		Success: true,
		Message: "healing state updated successfully",
	}, nil
}

// CheckDragonCanBattle checks if a dragon can participate in battles
func (s *DragonServiceServer) CheckDragonCanBattle(ctx context.Context, req *pb.CheckDragonCanBattleRequest) (*pb.CheckDragonCanBattleResponse, error) {
	dragonID, err := primitive.ObjectIDFromHex(req.DragonId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dragon ID: %v", err)
	}

	var dragon Dragon
	if err := DragonColl.FindOne(ctx, bson.M{"_id": dragonID}).Decode(&dragon); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "dragon not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get dragon: %v", err)
	}

	if !dragon.IsAlive {
		return &pb.CheckDragonCanBattleResponse{
			CanBattle:           false,
			Reason:              "dragon is not alive",
			HealingUntilSeconds: 0,
		}, nil
	}

	if dragon.IsHealing && dragon.HealingUntil != nil {
		now := time.Now()
        if now.Before(*dragon.HealingUntil) {
			return &pb.CheckDragonCanBattleResponse{
				CanBattle:           false,
				Reason:              "dragon is currently healing",
				HealingUntilSeconds: dragon.HealingUntil.Unix(),
			}, nil
		}
		// Healing expired, clear state
        _, _ = s.UpdateDragonHealingState(ctx, &pb.UpdateDragonHealingStateRequest{
			DragonId:           req.DragonId,
			IsHealing:          false,
			HealingUntilSeconds: 0,
		})
	}

	return &pb.CheckDragonCanBattleResponse{
		CanBattle:           true,
		Reason:              "",
		HealingUntilSeconds: 0,
	}, nil
}

