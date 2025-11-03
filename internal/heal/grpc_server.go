package heal

import (
	"context"

	pb "network-sec-micro/api/proto/heal"
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
	warriorID, err := parseWarriorID(req.WarriorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid warrior ID")
	}

	healType := HealType(req.HealType)
	if healType != HealTypeFull && healType != HealTypePartial {
		return nil, status.Error(codes.InvalidArgument, "invalid heal type. Use 'full' or 'partial'")
	}

	record, err := s.service.PurchaseHeal(ctx, warriorID, healType, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PurchaseHealResponse{
		Success:      true,
		Message:      "Healing applied successfully",
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

	records, err := s.service.GetHealingHistory(ctx, warriorID)
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

