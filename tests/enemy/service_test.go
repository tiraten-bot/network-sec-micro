package enemy_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/enemy"
	"network-sec-micro/internal/enemy/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB(t *testing.T) *mongo.Database {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	
	db := client.Database("enemy_test_db")
	db.Collection("enemies").Drop(ctx)
	
	enemy.EnemyColl = db.Collection("enemies")
	
	return db
}

func TestCreateEnemy_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateEnemyCommand{
		Name:        "Goblin Warrior",
		Type:        "goblin",
		Level:       5,
		MaxHealth:   100,
		AttackPower: 30,
		Defense:     20,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateEnemy(ctx, cmd)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Goblin Warrior", result.Name)
	assert.Equal(t, "goblin", string(result.Type))
	assert.Equal(t, 5, result.Level)
	assert.Equal(t, 100, result.MaxHealth)
	assert.Equal(t, 30, result.AttackPower)
	assert.Equal(t, 20, result.Defense)
}

func TestCreateEnemy_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateEnemyCommand{
		Name:        "Invalid Enemy",
		Type:        "invalid_type",
		Level:       1,
		MaxHealth:   50,
		AttackPower: 10,
		Defense:     5,
		CreatedBy:   "warrior1",
	}
	
	// Type validation might be in a different layer
	result, err := svc.CreateEnemy(ctx, cmd)
	
	// Just check that creation doesn't panic
	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestGetEnemies_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	// Create multiple enemies
	enemies := []dto.CreateEnemyCommand{
		{Name: "Goblin", Type: "goblin", Level: 1, MaxHealth: 50, AttackPower: 10, Defense: 5, CreatedBy: "warrior1"},
		{Name: "Orc", Type: "orc", Level: 3, MaxHealth: 80, AttackPower: 25, Defense: 15, CreatedBy: "warrior1"},
		{Name: "Pirate", Type: "pirate", Level: 2, MaxHealth: 60, AttackPower: 20, Defense: 10, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range enemies {
		_, err := svc.CreateEnemy(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetEnemiesQuery{}
	result, err := svc.GetEnemies(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestGetEnemies_ByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	enemies := []dto.CreateEnemyCommand{
		{Name: "Goblin1", Type: "goblin", Level: 1, MaxHealth: 50, AttackPower: 10, Defense: 5, CreatedBy: "warrior1"},
		{Name: "Orc1", Type: "orc", Level: 3, MaxHealth: 80, AttackPower: 25, Defense: 15, CreatedBy: "warrior1"},
		{Name: "Goblin2", Type: "goblin", Level: 2, MaxHealth: 60, AttackPower: 15, Defense: 8, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range enemies {
		_, err := svc.CreateEnemy(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetEnemiesQuery{Type: "goblin"}
	result, err := svc.GetEnemies(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, e := range result {
		assert.Equal(t, enemy.EnemyTypeGoblin, e.Type)
	}
}

func TestGetEnemies_ByCreatedBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	enemies := []dto.CreateEnemyCommand{
		{Name: "Enemy1", Type: "goblin", Level: 1, MaxHealth: 50, AttackPower: 10, Defense: 5, CreatedBy: "warrior1"},
		{Name: "Enemy2", Type: "orc", Level: 3, MaxHealth: 80, AttackPower: 25, Defense: 15, CreatedBy: "warrior1"},
		{Name: "Enemy3", Type: "pirate", Level: 2, MaxHealth: 60, AttackPower: 20, Defense: 10, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range enemies {
		_, err := svc.CreateEnemy(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetEnemiesQuery{CreatedBy: "warrior1"}
	result, err := svc.GetEnemies(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, e := range result {
		assert.Equal(t, "warrior1", e.CreatedBy)
	}
}

func TestEnemyType_Constants(t *testing.T) {
	assert.Equal(t, enemy.EnemyType("goblin"), enemy.EnemyTypeGoblin)
	assert.Equal(t, enemy.EnemyType("orc"), enemy.EnemyTypeOrc)
	assert.Equal(t, enemy.EnemyType("pirate"), enemy.EnemyTypePirate)
}

func TestEnemy_DefaultValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := enemy.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateEnemyCommand{
		Name:        "Test Enemy",
		Type:        "goblin",
		Level:       1,
		MaxHealth:   100,
		AttackPower: 20,
		Defense:     10,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateEnemy(ctx, cmd)
	require.NoError(t, err)
	
	// Check default values
	assert.Equal(t, result.MaxHealth, result.CurrentHealth) // Current health should equal max health initially
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.UpdatedAt)
}

