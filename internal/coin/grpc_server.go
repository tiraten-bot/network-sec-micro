package coin

import (
	"context"
	"errors"

	pb "network-sec-micro/api/proto/coin"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// CoinServiceServer implements the CoinService gRPC interface
type CoinServiceServer struct {
	pb.UnimplementedCoinServiceServer
	Service *Service
}

// NewCoinServiceServer creates a new coin gRPC server
func NewCoinServiceServer(service *Service) *CoinServiceServer {
	return &CoinServiceServer{
		Service: service,
	}
}

// Warrior represents a warrior from the warrior database
type Warrior struct {
	ID          uint
	CoinBalance int
}

// GetBalance returns warrior's coin balance from warrior database
func (s *CoinServiceServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", req.WarriorId).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "warrior not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get warrior: %v", err)
	}

	return &pb.GetBalanceResponse{
		WarriorId: uint32(warrior.ID),
		Balance:   int64(warrior.CoinBalance),
	}, nil
}

// DeductCoins deducts coins from warrior's balance
func (s *CoinServiceServer) DeductCoins(ctx context.Context, req *pb.DeductCoinsRequest) (*pb.DeductCoinsResponse, error) {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", req.WarriorId).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "warrior not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get warrior: %v", err)
	}

	balanceBefore := int64(warrior.CoinBalance)

	if balanceBefore < req.Amount {
		return &pb.DeductCoinsResponse{
			Success:       false,
			WarriorId:     uint32(warrior.ID),
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceBefore,
			Message:       "insufficient balance",
		}, nil
	}

	warrior.CoinBalance -= int(req.Amount)
	balanceAfter := int64(warrior.CoinBalance)

	if err := DB.Table("warriors").Where("id = ?", req.WarriorId).Update("coin_balance", warrior.CoinBalance).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update balance: %v", err)
	}

	// Create transaction record
	s.Service.CreateTransaction(
		warrior.ID,
		-req.Amount,
		TransactionTypeDeduct,
		req.Reason,
		balanceBefore,
		balanceAfter,
	)

	return &pb.DeductCoinsResponse{
		Success:       true,
		WarriorId:     uint32(warrior.ID),
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Message:       "coins deducted successfully",
	}, nil
}

// AddCoins adds coins to warrior's balance
func (s *CoinServiceServer) AddCoins(ctx context.Context, req *pb.AddCoinsRequest) (*pb.AddCoinsResponse, error) {
	var warrior Warrior
	if err := DB.Table("warriors").Where("id = ?", req.WarriorId).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "warrior not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get warrior: %v", err)
	}

	balanceBefore := int64(warrior.CoinBalance)
	warrior.CoinBalance += int(req.Amount)
	balanceAfter := int64(warrior.CoinBalance)

	if err := DB.Table("warriors").Where("id = ?", req.WarriorId).Update("coin_balance", warrior.CoinBalance).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update balance: %v", err)
	}

	// Create transaction record
	s.Service.CreateTransaction(
		warrior.ID,
		req.Amount,
		TransactionTypeAdd,
		req.Reason,
		balanceBefore,
		balanceAfter,
	)

	return &pb.AddCoinsResponse{
		Success:       true,
		WarriorId:     uint32(warrior.ID),
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Message:       "coins added successfully",
	}, nil
}

// TransferCoins transfers coins between warriors
func (s *CoinServiceServer) TransferCoins(ctx context.Context, req *pb.TransferCoinsRequest) (*pb.TransferCoinsResponse, error) {
	// Deduct from sender
	deductResp, err := s.DeductCoins(ctx, &pb.DeductCoinsRequest{
		WarriorId: req.FromWarriorId,
		Amount:    req.Amount,
		Reason:    "transfer_out: " + req.Reason,
	})
	if err != nil || !deductResp.Success {
		return &pb.TransferCoinsResponse{
			Success:       false,
			FromWarriorId: req.FromWarriorId,
			ToWarriorId:   req.ToWarriorId,
			Amount:        req.Amount,
			Message:       "failed to deduct coins from sender",
		}, nil
	}

	// Add to receiver
	addResp, err := s.AddCoins(ctx, &pb.AddCoinsRequest{
		WarriorId: req.ToWarriorId,
		Amount:    req.Amount,
		Reason:    "transfer_in: " + req.Reason,
	})
	if err != nil || !addResp.Success {
		// Rollback: add coins back to sender
		s.AddCoins(ctx, &pb.AddCoinsRequest{
			WarriorId: req.FromWarriorId,
			Amount:    req.Amount,
			Reason:    "rollback transfer_out",
		})

		return &pb.TransferCoinsResponse{
			Success:       false,
			FromWarriorId: req.FromWarriorId,
			ToWarriorId:   req.ToWarriorId,
			Amount:        req.Amount,
			Message:       "failed to add coins to receiver, transaction rolled back",
		}, nil
	}

	return &pb.TransferCoinsResponse{
		Success:       true,
		FromWarriorId: req.FromWarriorId,
		ToWarriorId:   req.ToWarriorId,
		Amount:        req.Amount,
		Message:       "coins transferred successfully",
	}, nil
}

// GetTransactionHistory returns transaction history for a warrior
func (s *CoinServiceServer) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	limit := int(req.Limit)
	if limit == 0 {
		limit = 50
	}
	offset := int(req.Offset)

	transactions, count, err := s.Service.GetTransactionsByWarrior(uint(req.WarriorId), limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTransactions := make([]*pb.Transaction, len(transactions))
	for i, tx := range transactions {
		protoTransactions[i] = &pb.Transaction{
			Id:               uint32(tx.ID),
			WarriorId:        uint32(tx.WarriorID),
			Amount:           tx.Amount,
			TransactionType: string(tx.TransactionType),
			Reason:           tx.Reason,
			CreatedAt:        timestamppb.New(tx.CreatedAt),
		}
	}

	return &pb.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		Total:        int32(count),
	}, nil
}
