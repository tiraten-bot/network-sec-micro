package arena

import (
	"context"
	"fmt"
	"log"
	"os"

	pbWarrior "network-sec-micro/api/proto/warrior"
    pbArenaSpell "network-sec-micro/api/proto/arenaspell"
    pbWeapon "network-sec-micro/api/proto/weapon"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn
var arenaspellGrpcClient pbArenaSpell.ArenaSpellServiceClient
var arenaspellGrpcConn *grpc.ClientConn
var weaponGrpcClient pbWeapon.WeaponServiceClient
var weaponGrpcConn *grpc.ClientConn

// InitWarriorClient initializes the gRPC client connection to warrior service
func InitWarriorClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("WARRIOR_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50052"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to warrior gRPC: %w", err)
	}

	warriorGrpcClient = pbWarrior.NewWarriorServiceClient(conn)
	warriorGrpcConn = conn

	log.Printf("Connected to Warrior gRPC service at %s", addr)
	return nil
}

// CloseWarriorClient closes the gRPC connection
func CloseWarriorClient() {
	if warriorGrpcConn != nil {
		warriorGrpcConn.Close()
	}
}

// InitArenaSpellClient initializes the gRPC client connection to arenaspell service
func InitArenaSpellClient(addr string) error {
    if addr == "" {
        addr = os.Getenv("ARENASPELL_GRPC_ADDR")
        if addr == "" {
            addr = "localhost:50056"
        }
    }

    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("failed to connect to arenaspell gRPC: %w", err)
    }

    arenaspellGrpcClient = pbArenaSpell.NewArenaSpellServiceClient(conn)
    arenaspellGrpcConn = conn
    log.Printf("Connected to ArenaSpell gRPC service at %s", addr)
    return nil
}

func CloseArenaSpellClient() {
    if arenaspellGrpcConn != nil {
        arenaspellGrpcConn.Close()
    }
}

// InitWeaponClient initializes the gRPC client connection to weapon service
func InitWeaponClient(addr string) error {
    if addr == "" {
        addr = os.Getenv("WEAPON_GRPC_ADDR")
        if addr == "" {
            addr = "localhost:50057"
        }
    }

    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("failed to connect to weapon gRPC: %w", err)
    }

    weaponGrpcClient = pbWeapon.NewWeaponServiceClient(conn)
    weaponGrpcConn = conn
    log.Printf("Connected to Weapon gRPC service at %s", addr)
    return nil
}

func CloseWeaponClient() {
    if weaponGrpcConn != nil { weaponGrpcConn.Close() }
}

// CastArenaSpellViaGRPC proxies to arenaspell.CastSpell
func CastArenaSpellViaGRPC(ctx context.Context, matchID string, spellType string, casterID uint, casterUsername, casterRole string) (int32, error) {
    if arenaspellGrpcClient == nil {
        return 0, fmt.Errorf("arenaspell gRPC client not initialized")
    }
    req := &pbArenaSpell.CastArenaSpellRequest{
        MatchId:        matchID,
        SpellType:      spellType,
        CasterUserId:   uint32(casterID),
        CasterUsername: casterUsername,
        CasterRole:     casterRole,
    }
    resp, err := arenaspellGrpcClient.CastSpell(ctx, req)
    if err != nil {
        return 0, err
    }
    if !resp.Success {
        return 0, fmt.Errorf(resp.Message)
    }
    return resp.AffectedCount, nil
}

// CalculateWarriorPowerViaWeapon queries weapon service to compute power
func CalculateWarriorPowerViaWeapon(ctx context.Context, username string) (totalPower int32, weaponCount int32, err error) {
    if weaponGrpcClient == nil {
        return 0, 0, fmt.Errorf("weapon gRPC client not initialized")
    }
    req := &pbWeapon.CalculateWarriorPowerRequest{WarriorUsername: username}
    resp, err := weaponGrpcClient.CalculateWarriorPower(ctx, req)
    if err != nil { return 0, 0, err }
    return resp.TotalPower, resp.WeaponCount, nil
}

// ListWeaponsByOwner fetches weapons for any owner type
func ListWeaponsByOwner(ctx context.Context, ownerType, ownerID string) ([]*pbWeapon.Weapon, error) {
    if weaponGrpcClient == nil { return nil, fmt.Errorf("weapon gRPC client not initialized") }
    resp, err := weaponGrpcClient.ListOwnerWeapons(ctx, &pbWeapon.ListOwnerWeaponsRequest{OwnerType: ownerType, OwnerId: ownerID})
    if err != nil { return nil, err }
    return resp.Weapons, nil
}

// ApplyWeaponWear reduces durability
func ApplyWeaponWear(ctx context.Context, weaponID string, wear int32) (*pbWeapon.ApplyWearResponse, error) {
    if weaponGrpcClient == nil { return nil, fmt.Errorf("weapon gRPC client not initialized") }
    return weaponGrpcClient.ApplyWear(ctx, &pbWeapon.ApplyWearRequest{WeaponId: weaponID, Wear: wear})
}

// GetWarriorByUsername gets warrior info via gRPC
func GetWarriorByUsername(ctx context.Context, username string) (*pbWarrior.Warrior, error) {
	if warriorGrpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.GetWarriorByUsernameRequest{
		Username: username,
	}

	resp, err := warriorGrpcClient.GetWarriorByUsername(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior by username: %w", err)
	}

	return resp.Warrior, nil
}

// GetWarriorByID gets warrior info by ID via gRPC
func GetWarriorByID(ctx context.Context, warriorID uint) (*pbWarrior.Warrior, error) {
	if warriorGrpcClient == nil {
		return nil, fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.GetWarriorByIDRequest{
		WarriorId: uint32(warriorID),
	}

	resp, err := warriorGrpcClient.GetWarriorByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get warrior: %w", err)
	}

	return resp.Warrior, nil
}

