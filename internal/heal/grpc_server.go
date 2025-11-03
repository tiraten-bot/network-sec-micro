package heal

import (
	"context"
	"fmt"

	pb "network-sec-micro/api/proto/heal"
	"network-sec-micro/internal/heal/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HealServiceServer implements the HealService gRPC server
type HealServiceServer struct {
	pb.UnimplementedHealServiceServer
	service *Service
}

// NewHealServiceServer creates a new gRPC server for heal service
func NewHealServiceServer(service *Service) *HealServiceServer {
	return &HealServiceServer{
		service: service,
	}
}

// PurchaseHeal handles heal purchase request
func (s *HealServiceServer) PurchaseHeal(ctx context.Context, req *pb.PurchaseHealRequest) (*pb.PurchaseHealResponse, error) {
	participantID := req.ParticipantId
	participantType := req.ParticipantType
	healType := HealType(req.HealType)
	participantRole := req.ParticipantRole

	// Default participant type to warrior for backward compatibility
	if participantType == "" {
		participantType = "warrior"
	}

	// If role not provided, get it from the participant's service
	if participantRole == "" {
		if participantType == "warrior" {
			warriorID, err := parseWarriorID(participantID)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid warrior ID")
			}
			warrior, err := GetWarriorByID(ctx, warriorID)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to get warrior role")
			}
			participantRole = warrior.Role
		} else if participantType == "dragon" {
			// Dragons have role "dragon"
			participantRole = "dragon"
		} else if participantType == "enemy" {
			// Enemies don't have roles, default to warrior package
			participantRole = "warrior"
		}
	}

	cmd := dto.PurchaseHealCommand{
		ParticipantID:   participantID,
		ParticipantType: participantType,
		HealType:        string(healType),
		BattleID:        "",
		ParticipantRole: participantRole,
	}
	record, err := s.service.PurchaseHeal(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PurchaseHealResponse{
		Success:      true,
		Message:      fmt.Sprintf("Healing started. Will complete in %d seconds", record.Duration),
		HealedAmount: int32(record.HealedAmount),
		NewHp:        int32(record.HPAfter),
		CoinsSpent:   int32(record.CoinsSpent),
	}, nil
}

// GetHealingHistory retrieves healing history
func (s *HealServiceServer) GetHealingHistory(ctx context.Context, req *pb.GetHealingHistoryRequest) (*pb.GetHealingHistoryResponse, error) {
	warriorID, err := parseWarriorID(req.WarriorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid warrior ID")
	}

	query := dto.GetHealingHistoryQuery{
		WarriorID: warriorID,
	}
	records, err := s.service.GetHealingHistory(ctx, query)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbRecords := make([]*pb.HealingRecord, 0, len(records))
	for _, r := range records {
		pbRecords = append(pbRecords, &pb.HealingRecord{
			Id:           r.ID,
			WarriorId:    req.WarriorId,
			HealType:     string(r.HealType),
			HealedAmount: int32(r.HealedAmount),
			CoinsSpent:   int32(r.CoinsSpent),
			CreatedAt:    timestamppb.New(r.CreatedAt),
		})
	}

	return &pb.GetHealingHistoryResponse{
		Records: pbRecords,
	}, nil
}

func parseWarriorID(idStr string) (uint, error) {
	// Simple uint parsing
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	return id, err
}

