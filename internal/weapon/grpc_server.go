package weapon

import (
	"context"

	pb "network-sec-micro/api/proto/weapon"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.mongodb.org/mongo-driver/bson"
)

// WeaponServiceServer implements the WeaponService gRPC interface
type WeaponServiceServer struct {
	pb.UnimplementedWeaponServiceServer
}

// NewWeaponServiceServer creates a new weapon gRPC server
func NewWeaponServiceServer() *WeaponServiceServer {
	return &WeaponServiceServer{}
}

// GetWeapon returns weapon details
func (s *WeaponServiceServer) GetWeapon(ctx context.Context, req *pb.GetWeaponRequest) (*pb.GetWeaponResponse, error) {
	var w Weapon
	if err := WeaponColl.FindOne(ctx, bson.M{"_id": req.WeaponId}).Decode(&w); err != nil {
		return nil, status.Errorf(codes.NotFound, "weapon not found")
	}

	return &pb.GetWeaponResponse{
		Weapon: &pb.Weapon{
			Id:          w.ID.Hex(),
			Name:        w.Name,
			Description: w.Description,
			Type:        string(w.Type),
			Damage:      int32(w.Damage),
			Price:       int32(w.Price),
			CreatedBy:   w.CreatedBy,
			OwnedBy:     w.OwnedBy,
		},
	}, nil
}

// CalculateWarriorPower calculates warrior's total power
func (s *WeaponServiceServer) CalculateWarriorPower(ctx context.Context, req *pb.CalculateWarriorPowerRequest) (*pb.CalculateWarriorPowerResponse, error) {
	// Get weapons owned by warrior
	weapons, err := GetWeaponsByOwner(ctx, req.WarriorUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get weapons: %v", err)
	}

	basePower := 100
	weaponBonus := 0

	for _, w := range weapons {
		weaponBonus += w.Damage
	}

	totalPower := basePower + weaponBonus

	return &pb.CalculateWarriorPowerResponse{
		WarriorUsername: req.WarriorUsername,
		BasePower:       int32(basePower),
		WeaponBonus:     int32(weaponBonus),
		TotalPower:      int32(totalPower),
		WeaponCount:     int32(len(weapons)),
	}, nil
}

// GetWeaponsByOwner gets weapons owned by a warrior
func GetWeaponsByOwner(ctx context.Context, warriorUsername string) ([]Weapon, error) {
	cursor, err := WeaponColl.Find(ctx, bson.M{"owned_by": warriorUsername})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var weapons []Weapon
	if err := cursor.All(ctx, &weapons); err != nil {
		return nil, err
	}

	return weapons, nil
}

