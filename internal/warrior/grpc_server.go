package warrior

import (
	"context"
	"log"
	"time"

	pb "network-sec-micro/api/proto/warrior"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WarriorServiceServer implements the WarriorService gRPC interface
type WarriorServiceServer struct {
	pb.UnimplementedWarriorServiceServer
	Service *Service
}

// NewWarriorServiceServer creates a new warrior gRPC server
func NewWarriorServiceServer(service *Service) *WarriorServiceServer {
	return &WarriorServiceServer{
		Service: service,
	}
}

// GetWarriorByUsername returns warrior by username
func (s *WarriorServiceServer) GetWarriorByUsername(ctx context.Context, req *pb.GetWarriorByUsernameRequest) (*pb.GetWarriorByUsernameResponse, error) {
	var w Warrior
	if err := DB.Where("username = ?", req.Username).First(&w).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "warrior not found: %v", err)
	}

	maxHP := w.MaxHP
	if maxHP == 0 {
		maxHP = w.TotalPower * 10 // Calculate if not set
		if maxHP < 100 {
			maxHP = 100
		}
	}
	return &pb.GetWarriorByUsernameResponse{
		Warrior: &pb.Warrior{
			Id:           uint32(w.ID),
			Username:     w.Username,
			Email:        w.Email,
			Role:         string(w.Role),
			CoinBalance:  int32(w.CoinBalance),
			TotalPower:   int32(w.TotalPower),
			WeaponCount:  int32(w.WeaponCount),
			CurrentHp:    int32(w.CurrentHP),
			MaxHp:        int32(maxHP),
			CreatedAt:    timestamppb.New(w.CreatedAt),
			UpdatedAt:    timestamppb.New(w.UpdatedAt),
		},
	}, nil
}

// GetWarriorByID returns warrior by ID
func (s *WarriorServiceServer) GetWarriorByID(ctx context.Context, req *pb.GetWarriorByIDRequest) (*pb.GetWarriorByIDResponse, error) {
	var w Warrior
	if err := DB.First(&w, req.WarriorId).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "warrior not found: %v", err)
	}

	maxHP := w.MaxHP
	if maxHP == 0 {
		maxHP = w.TotalPower * 10
		if maxHP < 100 {
			maxHP = 100
		}
	}
	return &pb.GetWarriorByIDResponse{
		Warrior: &pb.Warrior{
			Id:           uint32(w.ID),
			Username:     w.Username,
			Email:        w.Email,
			Role:         string(w.Role),
			CoinBalance:  int32(w.CoinBalance),
			TotalPower:   int32(w.TotalPower),
			WeaponCount:  int32(w.WeaponCount),
			CurrentHp:    int32(w.CurrentHP),
			MaxHp:        int32(maxHP),
			CreatedAt:    timestamppb.New(w.CreatedAt),
			UpdatedAt:    timestamppb.New(w.UpdatedAt),
		},
	}, nil
}

// UpdateWarriorPower updates warrior's power stats
func (s *WarriorServiceServer) UpdateWarriorPower(ctx context.Context, req *pb.UpdateWarriorPowerRequest) (*pb.UpdateWarriorPowerResponse, error) {
	if err := DB.Model(&Warrior{}).Where("id = ?", req.WarriorId).Updates(map[string]interface{}{
		"total_power":  req.TotalPower,
		"weapon_count": req.WeaponCount,
	}).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update power: %v", err)
	}

	return &pb.UpdateWarriorPowerResponse{
		Success: true,
		Message: "power updated successfully",
	}, nil
}

// UpdateWarriorHP updates warrior's HP (for healing)
func (s *WarriorServiceServer) UpdateWarriorHP(ctx context.Context, req *pb.UpdateWarriorHPRequest) (*pb.UpdateWarriorHPResponse, error) {
	var w Warrior
	if err := DB.First(&w, req.WarriorId).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "warrior not found: %v", err)
	}

	oldHP := w.CurrentHP

	// Update HP
	if err := DB.Model(&w).Update("current_hp", req.NewHp).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update HP: %v", err)
	}

	// Update MaxHP if not set (calculate from total_power)
	maxHP := w.MaxHP
	if maxHP == 0 {
		maxHP = w.TotalPower * 10
		if maxHP < 100 {
			maxHP = 100
		}
		if err := DB.Model(&w).Update("max_hp", maxHP).Error; err != nil {
			log.Printf("Warning: failed to update max_hp: %v", err)
		}
	}

	return &pb.UpdateWarriorHPResponse{
		Success: true,
		Message: "HP updated successfully",
		OldHp:   int32(oldHP),
		NewHp:   req.NewHp,
	}, nil
}

// UpdateWarriorHealingState updates warrior's healing state
func (s *WarriorServiceServer) UpdateWarriorHealingState(ctx context.Context, req *pb.UpdateWarriorHealingStateRequest) (*pb.UpdateWarriorHealingStateResponse, error) {
	var w Warrior
	if err := DB.First(&w, req.WarriorId).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "warrior not found: %v", err)
	}

	updates := map[string]interface{}{
		"is_healing": req.IsHealing,
	}

	if req.HealingUntilSeconds > 0 {
		healingUntil := time.Unix(req.HealingUntilSeconds, 0)
		updates["healing_until"] = healingUntil
	} else {
		updates["healing_until"] = nil
	}

	if err := DB.Model(&w).Updates(updates).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update healing state: %v", err)
	}

	return &pb.UpdateWarriorHealingStateResponse{
		Success: true,
		Message: "healing state updated successfully",
	}, nil
}

