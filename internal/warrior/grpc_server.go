package warrior

import (
	"context"

	pb "network-sec-micro/api/proto/warrior"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	return &pb.GetWarriorByUsernameResponse{
		Warrior: &pb.Warrior{
			Id:           uint32(w.ID),
			Username:     w.Username,
			Email:        w.Email,
			Role:         string(w.Role),
			CoinBalance:  int32(w.CoinBalance),
			TotalPower:   int32(w.TotalPower),
			WeaponCount:  int32(w.WeaponCount),
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

	return &pb.GetWarriorByIDResponse{
		Warrior: &pb.Warrior{
			Id:           uint32(w.ID),
			Username:     w.Username,
			Email:        w.Email,
			Role:         string(w.Role),
			CoinBalance:  int32(w.CoinBalance),
			TotalPower:   int32(w.TotalPower),
			WeaponCount:  int32(w.WeaponCount),
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

