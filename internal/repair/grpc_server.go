package repair

import (
    "context"
    "fmt"
    "time"

    pb "network-sec-micro/api/proto/repair"
    pbWeapon "network-sec-micro/api/proto/weapon"
    pbArmor "network-sec-micro/api/proto/armor"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcServer struct {
    pb.UnimplementedRepairServiceServer
    svc *Service
    weaponClient pbWeapon.WeaponServiceClient
    armorClient  pbArmor.ArmorServiceClient
}

func NewGrpcServer(svc *Service, weaponClient pbWeapon.WeaponServiceClient, armorClient pbArmor.ArmorServiceClient) *GrpcServer {
    return &GrpcServer{svc: svc, weaponClient: weaponClient, armorClient: armorClient}
}

func (g *GrpcServer) RepairWeapon(ctx context.Context, req *pb.RepairWeaponRequest) (*pb.RepairWeaponResponse, error) {
    if req.OwnerType == "" || req.OwnerId == "" || req.WeaponId == "" { return nil, status.Errorf(codes.InvalidArgument, "missing fields") }
    // Fetch weapon to compute cost
    gw, err := g.weaponClient.GetWeapon(ctx, &pbWeapon.GetWeaponRequest{WeaponId: req.WeaponId})
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "weapon not found") }
    cur := int(gw.Weapon.Durability); max := int(gw.Weapon.MaxDurability)
    if max == 0 { max = 100; if cur > max { max = cur } }
    // Use RBAC role for pricing (default to "warrior" if not provided)
    role := req.OwnerRole
    if role == "" { role = "warrior" }
    cost := g.svc.ComputeRepairCost(ctx, cur, max, role)
    if cost == 0 {
        return &pb.RepairWeaponResponse{Accepted: false, OrderId: "", Cost: 0, Status: "completed"}, nil
    }
    order, err := g.svc.CreateRepairOrder(ctx, req.OwnerType, req.OwnerId, req.WeaponId, "weapon", cost)
    if err != nil { return nil, status.Errorf(codes.Internal, "order create failed") }
    // Publish kafka event for coin deduction
    _ = PublishRepairEvent(ctx, req.OwnerType, req.OwnerId, cost, req.WeaponId, fmt.Sprintf("%d", order.ID))
    // Restore durability: apply negative wear to restore to max
    restoreAmount := max - cur
    if restoreAmount > 0 {
        _, _ = g.weaponClient.ApplyWear(ctx, &pbWeapon.ApplyWearRequest{WeaponId: req.WeaponId, Wear: int32(-restoreAmount)})
    }
    _ = g.svc.CompleteRepair(ctx, order.ID)
    return &pb.RepairWeaponResponse{Accepted: true, OrderId: fmt.Sprintf("%d", order.ID), Cost: int32(cost), Status: string(RepairStatusCompleted)}, nil
}

func (g *GrpcServer) RepairArmor(ctx context.Context, req *pb.RepairArmorRequest) (*pb.RepairArmorResponse, error) {
    if req.OwnerType == "" || req.OwnerId == "" || req.ArmorId == "" { return nil, status.Errorf(codes.InvalidArgument, "missing fields") }
    // Fetch armor to compute cost
    ga, err := g.armorClient.GetArmor(ctx, &pbArmor.GetArmorRequest{ArmorId: req.ArmorId})
    if err != nil { return nil, status.Errorf(codes.InvalidArgument, "armor not found") }
    cur := int(ga.Armor.Durability); max := int(ga.Armor.MaxDurability)
    if max == 0 { max = 100; if cur > max { max = cur } }
    // Use RBAC role for pricing (default to "warrior" if not provided)
    role := req.OwnerRole
    if role == "" { role = "warrior" }
    cost := g.svc.ComputeRepairCost(ctx, cur, max, role)
    if cost == 0 {
        return &pb.RepairArmorResponse{Accepted: false, OrderId: "", Cost: 0, Status: "completed"}, nil
    }
    order, err := g.svc.CreateRepairOrder(ctx, req.OwnerType, req.OwnerId, req.ArmorId, "armor", cost)
    if err != nil { return nil, status.Errorf(codes.Internal, "order create failed") }
    // Publish kafka event for coin deduction
    _ = PublishArmorRepairEvent(ctx, req.OwnerType, req.OwnerId, cost, req.ArmorId, fmt.Sprintf("%d", order.ID))
    // Restore durability: apply negative wear to restore to max
    restoreAmount := max - cur
    if restoreAmount > 0 {
        _, _ = g.armorClient.ApplyWear(ctx, &pbArmor.ApplyWearRequest{ArmorId: req.ArmorId, Wear: int32(-restoreAmount)})
    }
    _ = g.svc.CompleteRepair(ctx, order.ID)
    return &pb.RepairArmorResponse{Accepted: true, OrderId: fmt.Sprintf("%d", order.ID), Cost: int32(cost), Status: string(RepairStatusCompleted)}, nil
}

func (g *GrpcServer) GetRepairHistory(ctx context.Context, req *pb.GetRepairHistoryRequest) (*pb.GetRepairHistoryResponse, error) {
    if req.OwnerType == "" || req.OwnerId == "" { return nil, status.Errorf(codes.InvalidArgument, "missing fields") }
    orders, err := g.svc.ListOrders(ctx, req.OwnerType, req.OwnerId)
    if err != nil { return nil, status.Errorf(codes.Internal, "query failed") }
    out := make([]*pb.RepairOrderRecord, 0, len(orders))
    for _, o := range orders {
        rec := &pb.RepairOrderRecord{Id: fmt.Sprintf("%d", o.ID), OwnerType: o.OwnerType, OwnerId: o.OwnerID, WeaponId: o.WeaponID, Cost: int32(o.Cost), Status: string(o.Status)}
        rec.CreatedAt = timestamppb.New(o.CreatedAt)
        if o.CompletedAt != nil { rec.CompletedAt = timestamppb.New(*o.CompletedAt) }
        out = append(out, rec)
    }
    return &pb.GetRepairHistoryResponse{Orders: out}, nil
}


