package main

import (
    "log"
    "net"
    "os"

    pb "network-sec-micro/api/proto/repair"
    pbWeapon "network-sec-micro/api/proto/weapon"
    pbArmor "network-sec-micro/api/proto/armor"
    "network-sec-micro/internal/repair"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func getEnv(key, def string) string { if v := os.Getenv(key); v != "" { return v }; return def }

func main() {
    if err := repair.InitPostgres(); err != nil { log.Fatalf("db init error: %v", err) }
    // Wire DI for CQRS service
    svc, err := InitializeRepair()
    if err != nil {
        log.Printf("wire init failed, falling back: %v", err)
        svc = repair.NewService(repair.GetRepository())
    }

    // connect to weapon service
    waddr := getEnv("WEAPON_GRPC_ADDR", "localhost:50057")
    wconn, err := grpc.Dial(waddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { log.Fatalf("weapon grpc dial error: %v", err) }
    defer wconn.Close()
    wcli := pbWeapon.NewWeaponServiceClient(wconn)

    // connect to armor service
    aaddr := getEnv("ARMOR_GRPC_ADDR", "localhost:50059")
    aconn, err := grpc.Dial(aaddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { log.Fatalf("armor grpc dial error: %v", err) }
    defer aconn.Close()
    acli := pbArmor.NewArmorServiceClient(aconn)

    lis, err := net.Listen("tcp", ":50061")
    if err != nil { log.Fatalf("listen error: %v", err) }
    srv := grpc.NewServer()
    pb.RegisterRepairServiceServer(srv, repair.NewGrpcServer(svc, wcli, acli))
    log.Printf("repair service listening on %s", ":50061")
    if err := srv.Serve(lis); err != nil { log.Fatalf("serve error: %v", err) }
}


