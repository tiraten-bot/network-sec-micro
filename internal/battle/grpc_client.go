package battle

import (
	"context"
	"fmt"
	"log"
	"os"

	pbWarrior "network-sec-micro/api/proto/warrior"
	pbCoin "network-sec-micro/api/proto/coin"
	pbBattleSpell "network-sec-micro/api/proto/battlespell"
    pbWeapon "network-sec-micro/api/proto/weapon"
    pbArmor "network-sec-micro/api/proto/armor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn

var coinGrpcClient pbCoin.CoinServiceClient
var coinGrpcConn *grpc.ClientConn

var battlespellGrpcClient pbBattleSpell.BattleSpellServiceClient
var battlespellGrpcConn *grpc.ClientConn

var weaponGrpcClient pbWeapon.WeaponServiceClient
var weaponGrpcConn *grpc.ClientConn

var armorGrpcClient pbArmor.ArmorServiceClient
var armorGrpcConn *grpc.ClientConn

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

// InitCoinClient initializes the gRPC client connection to coin service
func InitCoinClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("COIN_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50051"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to coin gRPC: %w", err)
	}

	coinGrpcClient = pbCoin.NewCoinServiceClient(conn)
	coinGrpcConn = conn

	log.Printf("Connected to Coin gRPC service at %s", addr)
	return nil
}

// InitBattlespellClient initializes the gRPC client connection to battlespell service
func InitBattlespellClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("BATTLESPELL_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50054"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to battlespell gRPC: %w", err)
	}

	battlespellGrpcClient = pbBattleSpell.NewBattleSpellServiceClient(conn)
	battlespellGrpcConn = conn

	log.Printf("Connected to BattleSpell gRPC service at %s", addr)
	return nil
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

// CloseBattlespellClient closes the battlespell gRPC connection
func CloseBattlespellClient() {
	if battlespellGrpcConn != nil {
		battlespellGrpcConn.Close()
	}
}

// CloseWeaponClient closes the weapon gRPC connection
func CloseWeaponClient() {
    if weaponGrpcConn != nil {
        weaponGrpcConn.Close()
    }
}

// GetBattlespellClient returns the battlespell gRPC client
func GetBattlespellClient() pbBattleSpell.BattleSpellServiceClient {
	return battlespellGrpcClient
}

// CloseWarriorClient closes the gRPC connection
func CloseWarriorClient() {
	if warriorGrpcConn != nil {
		warriorGrpcConn.Close()
	}
}

// CloseCoinClient closes the coin gRPC connection
func CloseCoinClient() {
	if coinGrpcConn != nil {
		coinGrpcConn.Close()
	}
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

// GetWarriorClient returns the warrior gRPC client
func GetWarriorClient() pbWarrior.WarriorServiceClient {
	return warriorGrpcClient
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
    if weaponGrpcClient == nil {
        return nil, fmt.Errorf("weapon gRPC client not initialized")
    }
    resp, err := weaponGrpcClient.ListOwnerWeapons(ctx, &pbWeapon.ListOwnerWeaponsRequest{OwnerType: ownerType, OwnerId: ownerID})
    if err != nil { return nil, err }
    return resp.Weapons, nil
}

// ApplyWeaponWear reduces durability
func ApplyWeaponWear(ctx context.Context, weaponID string, wear int32) (*pbWeapon.ApplyWearResponse, error) {
    if weaponGrpcClient == nil { return nil, fmt.Errorf("weapon gRPC client not initialized") }
    return weaponGrpcClient.ApplyWear(ctx, &pbWeapon.ApplyWearRequest{WeaponId: weaponID, Wear: wear})
}

// AddCoins adds coins to warrior's balance via gRPC
func AddCoins(ctx context.Context, warriorID uint, amount int64, reason string) error {
	if coinGrpcClient == nil {
		return fmt.Errorf("coin gRPC client not initialized")
	}

	req := &pbCoin.AddCoinsRequest{
		WarriorId: uint32(warriorID),
		Amount:    amount,
		Reason:    reason,
	}

	resp, err := coinGrpcClient.AddCoins(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to add coins: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to add coins: %s", resp.Message)
	}

	log.Printf("Added %d coins to warrior %d. Balance: %d -> %d", amount, warriorID, resp.BalanceBefore, resp.BalanceAfter)
	return nil
}

// DeductCoins deducts coins from warrior's balance via gRPC
func DeductCoins(ctx context.Context, warriorID uint, amount int64, reason string) error {
	if coinGrpcClient == nil {
		return fmt.Errorf("coin gRPC client not initialized")
	}

	req := &pbCoin.DeductCoinsRequest{
		WarriorId: uint32(warriorID),
		Amount:    amount,
		Reason:    reason,
	}

	resp, err := coinGrpcClient.DeductCoins(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to deduct coins: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to deduct coins: %s", resp.Message)
	}

	log.Printf("Deducted %d coins from warrior %d. Balance: %d -> %d", amount, warriorID, resp.BalanceBefore, resp.BalanceAfter)
	return nil
}

