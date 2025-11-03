package arena

import (
	"context"

	pb "network-sec-micro/api/proto/arena"
	pbWarrior "network-sec-micro/api/proto/warrior"
)

// ArenaServiceServer implements the ArenaService gRPC interface
type ArenaServiceServer struct {
	pb.UnimplementedArenaServiceServer
}

// NewArenaServiceServer creates a new arena gRPC server
func NewArenaServiceServer() *ArenaServiceServer {
	return &ArenaServiceServer{}
}

// GetWarriorByUsername gets warrior info by username via Warrior service gRPC
func (s *ArenaServiceServer) GetWarriorByUsername(ctx context.Context, req *pb.GetWarriorByUsernameRequest) (*pb.GetWarriorByUsernameResponse, error) {
	warrior, err := GetWarriorByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.GetWarriorByUsernameResponse{
		Warrior: toProtoWarrior(warrior),
	}, nil
}

// GetWarriorByID gets warrior info by ID via Warrior service gRPC
func (s *ArenaServiceServer) GetWarriorByID(ctx context.Context, req *pb.GetWarriorByIDRequest) (*pb.GetWarriorByIDResponse, error) {
	warrior, err := GetWarriorByID(ctx, uint(req.WarriorId))
	if err != nil {
		return nil, err
	}

	return &pb.GetWarriorByIDResponse{
		Warrior: toProtoWarrior(warrior),
	}, nil
}

// toProtoWarrior converts warrior proto to arena proto warrior
func toProtoWarrior(warrior *pbWarrior.Warrior) *pb.Warrior {
	return &pb.Warrior{
		Id:          warrior.Id,
		Username:    warrior.Username,
		Email:       warrior.Email,
		Role:        warrior.Role,
		TotalPower:  warrior.TotalPower,
		AttackPower: warrior.AttackPower,
		Defense:     warrior.Defense,
		CoinBalance: warrior.CoinBalance,
		WeaponCount: warrior.WeaponCount,
	}
}

