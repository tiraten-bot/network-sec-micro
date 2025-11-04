package weapon

import (
	"context"
    "time"

	pb "network-sec-micro/api/proto/weapon"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
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
    oid, err := primitive.ObjectIDFromHex(req.WeaponId)
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "invalid weapon id") }
    if err := WeaponColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&w); err != nil {
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
            Durability:  int32(w.Durability),
            MaxDurability: int32(w.MaxDurability),
            IsBroken:    w.IsBroken,
            Owners: func() []*pb.OwnerRef {
                if len(w.Owners) == 0 { return nil }
                out := make([]*pb.OwnerRef, 0, len(w.Owners))
                for _, o := range w.Owners { out = append(out, &pb.OwnerRef{OwnerType: o.OwnerType, OwnerId: o.OwnerID}) }
                return out
            }(),
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

// ListOwnerWeapons lists weapons by generalized owner reference
func (s *WeaponServiceServer) ListOwnerWeapons(ctx context.Context, req *pb.ListOwnerWeaponsRequest) (*pb.ListOwnerWeaponsResponse, error) {
    filter := bson.M{}
    // Prefer new owners array; fallback to legacy for warriors
    if req.OwnerType != "" && req.OwnerId != "" {
        filter = bson.M{"owners": bson.M{"$elemMatch": bson.M{"owner_type": req.OwnerType, "owner_id": req.OwnerId}}}
        if req.OwnerType == "warrior" {
            // also include legacy field match
            filter = bson.M{"$or": []bson.M{
                {"owners": bson.M{"$elemMatch": bson.M{"owner_type": req.OwnerType, "owner_id": req.OwnerId}}},
                {"owned_by": req.OwnerId},
            }}
        }
    }
    cursor, err := WeaponColl.Find(ctx, filter)
    if err != nil { return nil, status.Errorf(codes.Internal, "query error: %v", err) }
    defer cursor.Close(ctx)
    var res []*pb.Weapon
    for cursor.Next(ctx) {
        var w Weapon
        if err := cursor.Decode(&w); err != nil { return nil, status.Errorf(codes.Internal, "decode error: %v", err) }
        res = append(res, &pb.Weapon{
            Id: w.ID.Hex(), Name: w.Name, Description: w.Description, Type: string(w.Type), Damage: int32(w.Damage), Price: int32(w.Price),
            CreatedBy: w.CreatedBy, OwnedBy: w.OwnedBy, Durability: int32(w.Durability), MaxDurability: int32(w.MaxDurability), IsBroken: w.IsBroken,
            Owners: func() []*pb.OwnerRef { if len(w.Owners)==0 {return nil}; out:=make([]*pb.OwnerRef,0,len(w.Owners)); for _,o:= range w.Owners { out = append(out, &pb.OwnerRef{OwnerType:o.OwnerType, OwnerId:o.OwnerID})}; return out }(),
        })
    }
    return &pb.ListOwnerWeaponsResponse{Weapons: res}, nil
}

// ApplyWear reduces durability and sets is_broken when needed
func (s *WeaponServiceServer) ApplyWear(ctx context.Context, req *pb.ApplyWearRequest) (*pb.ApplyWearResponse, error) {
    if req.Wear <= 0 { req.Wear = 1 }
    oid, err := primitive.ObjectIDFromHex(req.WeaponId)
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "invalid weapon id") }
    // Pull current values
    var w Weapon
    if err := WeaponColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&w); err != nil {
        return nil, status.Errorf(codes.NotFound, "weapon not found")
    }
    newDur := w.Durability - int(req.Wear)
    if newDur < 0 { newDur = 0 }
    isBroken := newDur == 0
    update := bson.M{"$set": bson.M{"durability": newDur, "is_broken": isBroken, "updated_at": time.Now()}}
    if _, err := WeaponColl.UpdateByID(ctx, oid, update); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to apply wear")
    }
    return &pb.ApplyWearResponse{WeaponId: req.WeaponId, Durability: int32(newDur), IsBroken: isBroken}, nil
}

