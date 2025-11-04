package armor

import (
    "context"
    "time"

    pb "network-sec-micro/api/proto/armor"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// ArmorServiceServer implements the ArmorService gRPC interface
type ArmorServiceServer struct {
    pb.UnimplementedArmorServiceServer
}

// NewArmorServiceServer creates a new armor gRPC server
func NewArmorServiceServer() *ArmorServiceServer { return &ArmorServiceServer{} }

// GetArmor returns armor details
func (s *ArmorServiceServer) GetArmor(ctx context.Context, req *pb.GetArmorRequest) (*pb.GetArmorResponse, error) {
    var a Armor
    oid, err := primitive.ObjectIDFromHex(req.ArmorId)
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "invalid armor id") }
    if err := ArmorColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&a); err != nil {
        return nil, status.Errorf(codes.NotFound, "armor not found")
    }
    return &pb.GetArmorResponse{ Armor: toProtoArmor(&a) }, nil
}

// CalculateDefense calculates owner's total defense
func (s *ArmorServiceServer) CalculateDefense(ctx context.Context, req *pb.CalculateDefenseRequest) (*pb.CalculateDefenseResponse, error) {
    armors, err := GetArmorsByOwner(ctx, req.OwnerType, req.OwnerId)
    if err != nil { return nil, status.Errorf(codes.Internal, "failed to get armors: %v", err) }
    base := 50
    bonus := 0
    for _, a := range armors { bonus += a.Defense }
    total := base + bonus
    return &pb.CalculateDefenseResponse{ OwnerType: req.OwnerType, OwnerId: req.OwnerId, BaseDefense: int32(base), ArmorBonus: int32(bonus), TotalDefense: int32(total), ArmorCount: int32(len(armors)) }, nil
}

// ListOwnerArmors lists armors by generalized owner reference
func (s *ArmorServiceServer) ListOwnerArmors(ctx context.Context, req *pb.ListOwnerArmorsRequest) (*pb.ListOwnerArmorsResponse, error) {
    filter := bson.M{}
    if req.OwnerType != "" && req.OwnerId != "" {
        filter = bson.M{"owners": bson.M{"$elemMatch": bson.M{"owner_type": req.OwnerType, "owner_id": req.OwnerId}}}
        if req.OwnerType == "warrior" {
            filter = bson.M{"$or": []bson.M{ {"owners": bson.M{"$elemMatch": bson.M{"owner_type": req.OwnerType, "owner_id": req.OwnerId}}}, {"owned_by": req.OwnerId} }}
        }
    }
    cur, err := ArmorColl.Find(ctx, filter)
    if err != nil { return nil, status.Errorf(codes.Internal, "query error: %v", err) }
    defer cur.Close(ctx)
    var res []*pb.Armor
    for cur.Next(ctx) {
        var a Armor
        if err := cur.Decode(&a); err != nil { return nil, status.Errorf(codes.Internal, "decode error") }
        res = append(res, toProtoArmor(&a))
    }
    return &pb.ListOwnerArmorsResponse{ Armors: res }, nil
}

// ApplyWear reduces durability and sets is_broken when needed
func (s *ArmorServiceServer) ApplyWear(ctx context.Context, req *pb.ApplyWearRequest) (*pb.ApplyWearResponse, error) {
    if req.Wear <= 0 { req.Wear = 1 }
    oid, err := primitive.ObjectIDFromHex(req.ArmorId)
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "invalid armor id") }
    var a Armor
    if err := ArmorColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&a); err != nil { return nil, status.Errorf(codes.NotFound, "armor not found") }
    newDur := a.Durability - int(req.Wear)
    if newDur < 0 { newDur = 0 }
    isBroken := newDur == 0
    upd := bson.M{"$set": bson.M{"durability": newDur, "is_broken": isBroken, "updated_at": time.Now()}}
    if _, err := ArmorColl.UpdateByID(ctx, oid, upd); err != nil { return nil, status.Errorf(codes.Internal, "failed to apply wear") }
    return &pb.ApplyWearResponse{ ArmorId: req.ArmorId, Durability: int32(newDur), IsBroken: isBroken }, nil
}

func toProtoArmor(a *Armor) *pb.Armor {
    return &pb.Armor{
        Id: a.ID.Hex(), Name: a.Name, Description: a.Description, Type: string(a.Type), Defense: int32(a.Defense), HpBonus: int32(a.HPBonus), Price: int32(a.Price), CreatedBy: a.CreatedBy,
        OwnedBy: a.OwnedBy,
        Durability: int32(a.Durability), MaxDurability: int32(a.MaxDurability), IsBroken: a.IsBroken,
        Owners: func() []*pb.OwnerRef { if len(a.Owners)==0 {return nil}; out:=make([]*pb.OwnerRef,0,len(a.Owners)); for _,o:= range a.Owners { out = append(out, &pb.OwnerRef{OwnerType:o.OwnerType, OwnerId:o.OwnerID})}; return out }(),
    }
}


