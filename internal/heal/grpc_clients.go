package heal

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pbWarrior "network-sec-micro/api/proto/warrior"
	pbCoin "network-sec-micro/api/proto/coin"
	pbDragon "network-sec-micro/api/proto/dragon"
	pbEnemy "network-sec-micro/api/proto/enemy"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var warriorGrpcClient pbWarrior.WarriorServiceClient
var warriorGrpcConn *grpc.ClientConn

var coinGrpcClient pbCoin.CoinServiceClient
var coinGrpcConn *grpc.ClientConn

var dragonGrpcClient pbDragon.DragonServiceClient
var dragonGrpcConn *grpc.ClientConn

var enemyGrpcClient pbEnemy.EnemyServiceClient
var enemyGrpcConn *grpc.ClientConn

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

// CloseCoinClient closes the coin gRPC connection
func CloseCoinClient() {
	if coinGrpcConn != nil {
		coinGrpcConn.Close()
	}
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

// UpdateWarriorHP updates warrior's HP via gRPC
func UpdateWarriorHP(ctx context.Context, warriorID uint, newHP int32) error {
	if warriorGrpcClient == nil {
		return fmt.Errorf("warrior gRPC client not initialized")
	}

	req := &pbWarrior.UpdateWarriorHPRequest{
		WarriorId: uint32(warriorID),
		NewHp:     newHP,
	}

	resp, err := warriorGrpcClient.UpdateWarriorHP(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update warrior HP: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update warrior HP: %s", resp.Message)
	}

	log.Printf("Updated warrior %d HP: %d -> %d", warriorID, resp.OldHp, resp.NewHp)
	return nil
}

// SetWarriorHealingState sets warrior's healing state (is_healing, healing_until)
func SetWarriorHealingState(ctx context.Context, warriorID uint, isHealing bool, healingUntil *time.Time) error {
	if warriorGrpcClient == nil {
		return fmt.Errorf("warrior gRPC client not initialized")
	}

	var healingUntilSeconds int64
	if healingUntil != nil {
		healingUntilSeconds = healingUntil.Unix()
	}

	req := &pbWarrior.UpdateWarriorHealingStateRequest{
		WarriorId:          uint32(warriorID),
		IsHealing:          isHealing,
		HealingUntilSeconds: healingUntilSeconds,
	}

	resp, err := warriorGrpcClient.UpdateWarriorHealingState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update warrior healing state: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update warrior healing state: %s", resp.Message)
	}

	if isHealing && healingUntil != nil {
		log.Printf("Warrior %d is now healing until %v", warriorID, healingUntil)
	} else {
		log.Printf("Warrior %d healing completed", warriorID)
	}
	return nil
}

// CheckWarriorHealingState checks if warrior is currently healing
func CheckWarriorHealingState(ctx context.Context, warriorID uint) (bool, *time.Time, error) {
	warrior, err := GetWarriorByID(ctx, warriorID)
	if err != nil {
		return false, nil, err
	}

	if !warrior.IsHealing || warrior.HealingUntilSeconds == 0 {
		return false, nil, nil
	}

	healingUntil := time.Unix(warrior.HealingUntilSeconds, 0)
	now := time.Now()
	if now.After(healingUntil) {
		// Healing time passed, clear state
		_ = SetWarriorHealingState(ctx, warriorID, false, nil)
		return false, nil, nil
	}

	return true, &healingUntil, nil
}

// ==================== DRAGON gRPC CLIENT ====================

// InitDragonClient initializes the gRPC client connection to dragon service
func InitDragonClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("DRAGON_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50059"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to dragon gRPC: %w", err)
	}

	dragonGrpcClient = pbDragon.NewDragonServiceClient(conn)
	dragonGrpcConn = conn

	log.Printf("Connected to Dragon gRPC service at %s", addr)
	return nil
}

// CloseDragonClient closes the dragon gRPC connection
func CloseDragonClient() {
	if dragonGrpcConn != nil {
		dragonGrpcConn.Close()
	}
}

// GetDragonByID gets dragon info by ID via gRPC
func GetDragonByID(ctx context.Context, dragonID string) (*pbDragon.Dragon, error) {
	if dragonGrpcClient == nil {
		return nil, fmt.Errorf("dragon gRPC client not initialized")
	}

	req := &pbDragon.GetDragonByIDRequest{
		DragonId: dragonID,
	}

	resp, err := dragonGrpcClient.GetDragonByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get dragon: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get dragon: %s", resp.Message)
	}

	return resp.Dragon, nil
}

// UpdateDragonHP updates dragon's HP via gRPC
func UpdateDragonHP(ctx context.Context, dragonID string, newHP int32) error {
	if dragonGrpcClient == nil {
		return fmt.Errorf("dragon gRPC client not initialized")
	}

	req := &pbDragon.UpdateDragonHPRequest{
		DragonId: dragonID,
		NewHp:   newHP,
	}

	resp, err := dragonGrpcClient.UpdateDragonHP(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update dragon HP: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update dragon HP: %s", resp.Message)
	}

	log.Printf("Updated dragon %s HP to %d", dragonID, resp.CurrentHp)
	return nil
}

// SetDragonHealingState sets dragon's healing state
func SetDragonHealingState(ctx context.Context, dragonID string, isHealing bool, healingUntil *time.Time) error {
	if dragonGrpcClient == nil {
		return fmt.Errorf("dragon gRPC client not initialized")
	}

	var healingUntilSeconds int64
	if healingUntil != nil {
		healingUntilSeconds = healingUntil.Unix()
	}

	req := &pbDragon.UpdateDragonHealingStateRequest{
		DragonId:           dragonID,
		IsHealing:          isHealing,
		HealingUntilSeconds: healingUntilSeconds,
	}

	resp, err := dragonGrpcClient.UpdateDragonHealingState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update dragon healing state: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update dragon healing state: %s", resp.Message)
	}

	return nil
}

// CheckDragonHealingState checks if dragon is currently healing
func CheckDragonHealingState(ctx context.Context, dragonID string) (bool, *time.Time, error) {
	dragon, err := GetDragonByID(ctx, dragonID)
	if err != nil {
		return false, nil, err
	}

	if !dragon.IsHealing || dragon.HealingUntilSeconds == 0 {
		return false, nil, nil
	}

	healingUntil := time.Unix(dragon.HealingUntilSeconds, 0)
	now := time.Now()
	if now.After(healingUntil) {
		_ = SetDragonHealingState(ctx, dragonID, false, nil)
		return false, nil, nil
	}

	return true, &healingUntil, nil
}

// ==================== ENEMY gRPC CLIENT ====================

// InitEnemyClient initializes the gRPC client connection to enemy service
func InitEnemyClient(addr string) error {
	if addr == "" {
		addr = os.Getenv("ENEMY_GRPC_ADDR")
		if addr == "" {
			addr = "localhost:50060"
		}
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to enemy gRPC: %w", err)
	}

	enemyGrpcClient = pbEnemy.NewEnemyServiceClient(conn)
	enemyGrpcConn = conn

	log.Printf("Connected to Enemy gRPC service at %s", addr)
	return nil
}

// CloseEnemyClient closes the enemy gRPC connection
func CloseEnemyClient() {
	if enemyGrpcConn != nil {
		enemyGrpcConn.Close()
	}
}

// GetEnemyByID gets enemy info by ID via gRPC
func GetEnemyByID(ctx context.Context, enemyID string) (*pbEnemy.Enemy, error) {
	if enemyGrpcClient == nil {
		return nil, fmt.Errorf("enemy gRPC client not initialized")
	}

	req := &pbEnemy.GetEnemyByIDRequest{
		EnemyId: enemyID,
	}

	resp, err := enemyGrpcClient.GetEnemyByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get enemy: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get enemy: %s", resp.Message)
	}

	return resp.Enemy, nil
}

// UpdateEnemyHP updates enemy's HP via gRPC
func UpdateEnemyHP(ctx context.Context, enemyID string, newHP int32) error {
	if enemyGrpcClient == nil {
		return fmt.Errorf("enemy gRPC client not initialized")
	}

	req := &pbEnemy.UpdateEnemyHPRequest{
		EnemyId: enemyID,
		NewHp:  newHP,
	}

	resp, err := enemyGrpcClient.UpdateEnemyHP(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update enemy HP: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update enemy HP: %s", resp.Message)
	}

	log.Printf("Updated enemy %s HP to %d", enemyID, resp.CurrentHp)
	return nil
}

// SetEnemyHealingState sets enemy's healing state
func SetEnemyHealingState(ctx context.Context, enemyID string, isHealing bool, healingUntil *time.Time) error {
	if enemyGrpcClient == nil {
		return fmt.Errorf("enemy gRPC client not initialized")
	}

	var healingUntilSeconds int64
	if healingUntil != nil {
		healingUntilSeconds = healingUntil.Unix()
	}

	req := &pbEnemy.UpdateEnemyHealingStateRequest{
		EnemyId:            enemyID,
		IsHealing:          isHealing,
		HealingUntilSeconds: healingUntilSeconds,
	}

	resp, err := enemyGrpcClient.UpdateEnemyHealingState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update enemy healing state: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update enemy healing state: %s", resp.Message)
	}

	return nil
}

// CheckEnemyHealingState checks if enemy is currently healing
func CheckEnemyHealingState(ctx context.Context, enemyID string) (bool, *time.Time, error) {
	enemy, err := GetEnemyByID(ctx, enemyID)
	if err != nil {
		return false, nil, err
	}

	if !enemy.IsHealing || enemy.HealingUntilSeconds == 0 {
		return false, nil, nil
	}

	healingUntil := time.Unix(enemy.HealingUntilSeconds, 0)
	now := time.Now()
	if now.After(healingUntil) {
		_ = SetEnemyHealingState(ctx, enemyID, false, nil)
		return false, nil, nil
	}

	return true, &healingUntil, nil
}

// DeductCoinsForParticipant deducts coins for a participant (warrior only for now)
func DeductCoinsForParticipant(ctx context.Context, participantID string, participantType string, amount int64, reason string) error {
	if participantType != "warrior" {
		// Dragons and enemies don't have coin balances (they are NPCs)
		// For now, we'll allow free healing for dragons/enemies
		// In the future, we might charge the creator's balance
		log.Printf("Skipping coin deduction for %s (type: %s) - NPCs don't have coin balances", participantID, participantType)
		return nil
	}

	warriorID, err := strconv.ParseUint(participantID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid warrior ID: %w", err)
	}

	return DeductCoins(ctx, uint(warriorID), amount, reason)
}

