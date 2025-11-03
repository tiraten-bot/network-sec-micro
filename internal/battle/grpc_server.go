package battle

import (
    "context"
    "fmt"
    "time"

    pb "network-sec-micro/api/proto/battle"
    "network-sec-micro/internal/battle/dto"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// BattleServiceServer implements the BattleService gRPC server
type BattleServiceServer struct {
	pb.UnimplementedBattleServiceServer
	service *Service
}

// NewBattleServiceServer creates a new gRPC server for battle service
func NewBattleServiceServer(service *Service) *BattleServiceServer {
	return &BattleServiceServer{
		service: service,
	}
}

// GetBattleByID gets battle by ID
func (s *BattleServiceServer) GetBattleByID(ctx context.Context, req *pb.GetBattleByIDRequest) (*pb.GetBattleByIDResponse, error) {
    if _, err := primitive.ObjectIDFromHex(req.BattleId); err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid battle ID")
    }

    battle, err := GetRepository().GetBattleByID(ctx, req.BattleId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "battle not found")
    }

    pbBattle := convertBattleToProto(battle)
	return &pb.GetBattleByIDResponse{
		Battle: pbBattle,
	}, nil
}

// GetActiveBattles gets active battles
func (s *BattleServiceServer) GetActiveBattles(ctx context.Context, req *pb.GetActiveBattlesRequest) (*pb.GetActiveBattlesResponse, error) {
	filter := bson.M{
		"status": BattleStatusInProgress,
	}

	if req.Side != "" {
		// Filter by side - would need to check participants
		// For simplicity, returning all active battles
	}

	cursor, err := BattleColl.Find(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find battles: %v", err))
	}
	defer cursor.Close(ctx)

	var battles []*pb.Battle
	for cursor.Next(ctx) {
		var battle Battle
		if err := cursor.Decode(&battle); err != nil {
			continue
		}
		pbBattle := convertBattleToProto(&battle)
		battles = append(battles, pbBattle)
	}

	return &pb.GetActiveBattlesResponse{
		Battles: battles,
	}, nil
}

// GetBattleParticipants gets battle participants
func (s *BattleServiceServer) GetBattleParticipants(ctx context.Context, req *pb.GetBattleParticipantsRequest) (*pb.GetBattleParticipantsResponse, error) {
	battleID, err := primitive.ObjectIDFromHex(req.BattleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid battle ID")
	}

	filter := primitive.M{
		"battle_id": battleID,
	}

	if req.Side != "" {
		filter["side"] = req.Side
	}

	cursor, err := BattleParticipantColl.Find(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find participants: %v", err))
	}
	defer cursor.Close(ctx)

	var participants []*pb.BattleParticipant
	for cursor.Next(ctx) {
		var participant BattleParticipant
		if err := cursor.Decode(&participant); err != nil {
			continue
		}
		pbParticipant := convertParticipantToProto(&participant)
		participants = append(participants, pbParticipant)
	}

	return &pb.GetBattleParticipantsResponse{
		Participants: participants,
	}, nil
}

// UpdateParticipantStats updates participant stats
func (s *BattleServiceServer) UpdateParticipantStats(ctx context.Context, req *pb.UpdateParticipantStatsRequest) (*pb.UpdateParticipantStatsResponse, error) {
	battleID, err := primitive.ObjectIDFromHex(req.BattleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid battle ID")
	}

	updateData := bson.M{
		"hp":          req.Hp,
		"max_hp":      req.MaxHp,
		"attack_power": req.AttackPower,
		"defense":     req.Defense,
		"is_alive":    req.IsAlive,
		"updated_at":  time.Now(),
	}

	_, err = BattleParticipantColl.UpdateOne(ctx, bson.M{
		"battle_id":      battleID,
		"participant_id": req.ParticipantId,
	}, bson.M{"$set": updateData})

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update participant: %v", err))
	}

	return &pb.UpdateParticipantStatsResponse{
		Success: true,
		Message: "participant stats updated successfully",
	}, nil
}

// CastSpell casts a spell via gRPC - delegates to battlespell service
func (s *BattleServiceServer) CastSpell(ctx context.Context, req *pb.CastSpellRequest) (*pb.CastSpellResponse, error) {
	// Battle service now delegates spell casting to battlespell service
	// This endpoint can be removed or kept for backward compatibility
	return &pb.CastSpellResponse{
		Success: false,
		Message: "spell casting has been moved to battlespell service, please use battlespell service gRPC endpoint",
	}, nil
}

// Helper functions to convert between internal models and proto
func convertBattleToProto(b *Battle) *pb.Battle {
	pbBattle := &pb.Battle{
		Id:                    b.ID.Hex(),
		BattleType:            string(b.BattleType),
		LightSideName:         b.LightSideName,
		DarkSideName:          b.DarkSideName,
		CurrentTurn:           int32(b.CurrentTurn),
		CurrentParticipantIndex: int32(b.CurrentParticipantIndex),
		MaxTurns:              int32(b.MaxTurns),
		Status:                string(b.Status),
		CreatedBy:             b.CreatedBy,
		CreatedAt:             timestamppb.New(b.CreatedAt),
		UpdatedAt:             timestamppb.New(b.UpdatedAt),
	}

	if b.Result != "" {
		pbBattle.Result = string(b.Result)
	}
	if b.WinnerSide != "" {
		pbBattle.WinnerSide = string(b.WinnerSide)
	}
	if b.StartedAt != nil {
		pbBattle.StartedAt = timestamppb.New(*b.StartedAt)
	}
	if b.CompletedAt != nil {
		pbBattle.CompletedAt = timestamppb.New(*b.CompletedAt)
	}

	return pbBattle
}

func convertParticipantToProto(p *BattleParticipant) *pb.BattleParticipant {
	pbParticipant := &pb.BattleParticipant{
		Id:            p.ID.Hex(),
		BattleId:     p.BattleID.Hex(),
		ParticipantId: p.ParticipantID,
		Name:          p.Name,
		Type:          string(p.Type),
		Side:          string(p.Side),
		Hp:            int32(p.HP),
		MaxHp:         int32(p.MaxHP),
		AttackPower:   int32(p.AttackPower),
		Defense:       int32(p.Defense),
		IsAlive:       p.IsAlive,
		IsDefeated:    p.IsDefeated,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}

	if p.DefeatedAt != nil {
		pbParticipant.DefeatedAt = timestamppb.New(*p.DefeatedAt)
	}

	return pbParticipant
}

