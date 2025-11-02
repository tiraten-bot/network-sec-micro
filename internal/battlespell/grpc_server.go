package battlespell

import (
	"context"
	"fmt"

	pb "network-sec-micro/api/proto/battlespell"
	"network-sec-micro/internal/battlespell/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BattleSpellServiceServer implements the BattleSpellService gRPC server
type BattleSpellServiceServer struct {
	pb.UnimplementedBattleSpellServiceServer
	service *Service
}

// NewBattleSpellServiceServer creates a new gRPC server for battlespell service
func NewBattleSpellServiceServer(service *Service) *BattleSpellServiceServer {
	return &BattleSpellServiceServer{
		service: service,
	}
}

// CastSpell casts a spell via gRPC
func (s *BattleSpellServiceServer) CastSpell(ctx context.Context, req *pb.CastSpellRequest) (*pb.CastSpellResponse, error) {
	cmd := dto.CastSpellCommand{
		BattleID:            req.BattleId,
		SpellType:           req.SpellType,
		CasterUsername:      req.CasterUsername,
		CasterUserID:        req.CasterUserId,
		CasterRole:          req.CasterRole,
		TargetDragonID:      req.TargetDragonId,
		TargetDarkEmperorID: req.TargetDarkEmperorId,
	}

	affectedCount, err := s.service.CastSpell(ctx, cmd)
	if err != nil {
		return &pb.CastSpellResponse{
			Success:      false,
			Message:      err.Error(),
			AffectedCount: 0,
		}, nil
	}

	return &pb.CastSpellResponse{
		Success:       true,
		Message:       fmt.Sprintf("spell %s cast successfully", req.SpellType),
		AffectedCount: int32(affectedCount),
	}, nil
}

// GetActiveSpells gets active spells for a battle
func (s *BattleSpellServiceServer) GetActiveSpells(ctx context.Context, req *pb.GetActiveSpellsRequest) (*pb.GetActiveSpellsResponse, error) {
	battleID, err := primitive.ObjectIDFromHex(req.BattleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid battle ID")
	}

	// Get active spells from database
	cursor, err := SpellColl.Find(ctx, map[string]interface{}{
		"battle_id": battleID,
		"is_active": true,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find spells: %v", err))
	}
	defer cursor.Close(ctx)

	var spells []*pb.Spell
	for cursor.Next(ctx) {
		var spell Spell
		if err := cursor.Decode(&spell); err != nil {
			continue
		}
		pbSpell := convertSpellToProto(&spell)
		spells = append(spells, pbSpell)
	}

	return &pb.GetActiveSpellsResponse{
		Spells: spells,
	}, nil
}

// TriggerWraithOfDragon triggers wraith effect
func (s *BattleSpellServiceServer) TriggerWraithOfDragon(ctx context.Context, req *pb.TriggerWraithOfDragonRequest) (*pb.TriggerWraithOfDragonResponse, error) {
	battleID, err := primitive.ObjectIDFromHex(req.BattleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid battle ID")
	}

	warriorID, err := s.service.TriggerWraithOfDragon(ctx, battleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get wraith count
	var spell Spell
	err = SpellColl.FindOne(ctx, map[string]interface{}{
		"battle_id":  battleID,
		"spell_type": SpellWraithOfDragon,
		"is_active": true,
	}).Decode(&spell)

	wraithCount := int32(0)
	if err == nil {
		wraithCount = int32(spell.WraithCount)
	}

	return &pb.TriggerWraithOfDragonResponse{
		Triggered:          warriorID != "",
		DestroyedWarriorId: warriorID,
		WraithCount:        wraithCount,
		Message:           fmt.Sprintf("Wraith triggered: %d/25", wraithCount),
	}, nil
}

// Helper function to convert Spell to proto
func convertSpellToProto(s *Spell) *pb.Spell {
	pbSpell := &pb.Spell{
		Id:               s.ID.Hex(),
		BattleId:         s.BattleID.Hex(),
		SpellType:        string(s.SpellType),
		Side:             string(s.Side),
		CasterUsername:   s.CasterUsername,
		CasterUserId:     s.CasterUserID,
		CasterRole:       s.CasterRole,
		StackCount:       int32(s.StackCount),
		WraithCount:      int32(s.WraithCount),
		IsActive:         s.IsActive,
		CastAt:           timestamppb.New(s.CastAt),
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}

	if s.TargetDragonID != "" {
		pbSpell.TargetDragonId = s.TargetDragonID
	}
	if s.TargetDarkEmperorID != "" {
		pbSpell.TargetDarkEmperorId = s.TargetDarkEmperorID
	}

	return pbSpell
}

