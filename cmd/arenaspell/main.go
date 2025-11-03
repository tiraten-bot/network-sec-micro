package main

import (
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"

    pb "network-sec-micro/api/proto/arenaspell"
    arenaspell "network-sec-micro/internal/arenaspell"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
)

func main() {
    if err := arenaspell.InitDatabase(); err != nil {
        log.Fatalf("arenaspell db init failed: %v", err)
    }

    _, h, grpcSrv, err := InitializeApp()
    if err != nil { log.Fatalf("arenaspell wire init failed: %v", err) }

    // gRPC server
    go func() {
        port := os.Getenv("ARENASPELL_GRPC_PORT")
        if port == "" { port = "50056" }
        lis, err := net.Listen("tcp", ":"+port)
        if err != nil { log.Fatalf("arenaspell grpc listen failed: %v", err) }
        s := grpc.NewServer()
        pb.RegisterArenaSpellServiceServer(s, grpcSrv)
        log.Printf("ArenaSpell gRPC starting on %s", port)
        if err := s.Serve(lis); err != nil { log.Fatalf("arenaspell grpc serve failed: %v", err) }
    }()

    // HTTP server
    r := gin.Default()
    arenaspell.SetupRoutes(r, h)
    httpAddr := os.Getenv("ARENASPELL_HTTP_ADDR")
    if httpAddr == "" { httpAddr = ":8088" }
    go func() {
        log.Printf("ArenaSpell HTTP starting on %s", httpAddr)
        if err := r.Run(httpAddr); err != nil { log.Fatalf("arenaspell http failed: %v", err) }
    }()

    // Wait for signal
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
    <-sig
    log.Println("arenaspell shutting down")
}


