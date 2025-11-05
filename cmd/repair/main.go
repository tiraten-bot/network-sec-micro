package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "network-sec-micro/api/proto/repair"
	pbWeapon "network-sec-micro/api/proto/weapon"
	pbArmor "network-sec-micro/api/proto/armor"
	"network-sec-micro/internal/repair"
	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

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

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start metrics server
	metricsPort := getEnv("METRICS_PORT", "8082")
	healthHandler := health.NewHandler(&health.DatabaseChecker{DB: repair.GetDB(), DBName: "postgres"})
	go func() {
		if err := metrics.StartMetricsServerWithContext(ctx, metricsPort, healthHandler); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50061")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterRepairServiceServer(srv, repair.NewGrpcServer(svc, wcli, acli))

	log.Printf("repair service listening on %s", ":50061")
	log.Printf("repair metrics server listening on %s", metricsPort)

	// Start gRPC server in goroutine
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, gracefully shutting down...")
	cancel()
	srv.GracefulStop()
	log.Println("Repair service stopped")
}


