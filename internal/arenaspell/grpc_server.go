package arenaspell

import (
    "context"
    "fmt"

    pb "network-sec-micro/api/proto/arenaspell"
    "network-sec-micro/internal/arenaspell/dto"
)

// ArenaSpellServiceServer implements the ArenaSpellService gRPC server
type ArenaSpellServiceServer struct {
    pb.UnimplementedArenaSpellServiceServer
    service *Service
}

func NewArenaSpellServiceServer(s *Service) *ArenaSpellServiceServer { return &ArenaSpellServiceServer{service: s} }

func (s *ArenaSpellServiceServer) CastSpell(ctx context.Context, req *pb.CastArenaSpellRequest) (*pb.CastArenaSpellResponse, error) {
    cmd := dto.CastArenaSpellCommand{
        MatchID:        req.MatchId,
        SpellType:      req.SpellType,
        CasterUserID:   uint(req.CasterUserId),
        CasterUsername: req.CasterUsername,
        CasterRole:     req.CasterRole,
    }

    affected, err := s.service.CastSpell(ctx, cmd)
    if err != nil {
        return &pb.CastArenaSpellResponse{Success: false, Message: err.Error(), AffectedCount: 0}, nil
    }
    return &pb.CastArenaSpellResponse{Success: true, Message: fmt.Sprintf("spell %s cast successfully", req.SpellType), AffectedCount: int32(affected)}, nil
}


