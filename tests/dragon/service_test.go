package dragon_test

import (
	"context"
	"testing"

	"network-sec-micro/internal/dragon"
	"network-sec-micro/internal/dragon/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB(t *testing.T) *mongo.Database {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	
	db := client.Database("dragon_test_db")
	db.Collection("dragons").Drop(ctx)
	
	dragon.DragonColl = db.Collection("dragons")
	
	return db
}

func TestCreateDragon_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := dragon.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateDragonCommand{
		Name:        "Fire Dragon",
		Type:        "fire",
		Level:       10,
		MaxHealth:   500,
		AttackPower: 100,
		Defense:     80,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateDragon(ctx, cmd)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Fire Dragon", result.Name)
	assert.Equal(t, "fire", string(result.Type))
	assert.Equal(t, 10, result.Level)
	assert.Equal(t, 500, result.MaxHealth)
}

func TestGetDragons_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := dragon.NewService()
	ctx := context.Background()
	
	dragons := []dto.CreateDragonCommand{
		{Name: "Fire Dragon", Type: "fire", Level: 10, MaxHealth: 500, AttackPower: 100, Defense: 80, CreatedBy: "warrior1"},
		{Name: "Ice Dragon", Type: "ice", Level: 12, MaxHealth: 600, AttackPower: 120, Defense: 90, CreatedBy: "warrior1"},
		{Name: "Lightning Dragon", Type: "lightning", Level: 15, MaxHealth: 700, AttackPower: 150, Defense: 100, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range dragons {
		_, err := svc.CreateDragon(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetDragonsQuery{}
	result, err := svc.GetDragons(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestGetDragons_ByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := dragon.NewService()
	ctx := context.Background()
	
	dragons := []dto.CreateDragonCommand{
		{Name: "Fire Dragon 1", Type: "fire", Level: 10, MaxHealth: 500, AttackPower: 100, Defense: 80, CreatedBy: "warrior1"},
		{Name: "Ice Dragon", Type: "ice", Level: 12, MaxHealth: 600, AttackPower: 120, Defense: 90, CreatedBy: "warrior1"},
		{Name: "Fire Dragon 2", Type: "fire", Level: 15, MaxHealth: 700, AttackPower: 150, Defense: 100, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range dragons {
		_, err := svc.CreateDragon(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetDragonsQuery{Type: "fire"}
	result, err := svc.GetDragons(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, d := range result {
		assert.Equal(t, dragon.DragonTypeFire, d.Type)
	}
}

func TestDragonType_Constants(t *testing.T) {
	assert.Equal(t, dragon.DragonType("fire"), dragon.DragonTypeFire)
	assert.Equal(t, dragon.DragonType("ice"), dragon.DragonTypeIce)
	assert.Equal(t, dragon.DragonType("lightning"), dragon.DragonTypeLightning)
	assert.Equal(t, dragon.DragonType("shadow"), dragon.DragonTypeShadow)
}

func TestDragon_RevivalMechanics(t *testing.T) {
	// Test dragon revival count logic
	dragon := dragon.Dragon{
		RevivalCount: 0,
		MaxRevivals:  3,
	}
	
	// Dragon should be alive initially
	assert.True(t, dragon.RevivalCount < dragon.MaxRevivals)
	
	// After 3 deaths, dragon should be permanently dead
	dragon.RevivalCount = 3
	assert.False(t, dragon.RevivalCount < dragon.MaxRevivals)
}

func TestDragon_DefaultValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := dragon.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateDragonCommand{
		Name:        "Test Dragon",
		Type:        "fire",
		Level:       5,
		MaxHealth:   300,
		AttackPower: 50,
		Defense:     40,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateDragon(ctx, cmd)
	require.NoError(t, err)
	
	// Check default values
	assert.Equal(t, result.MaxHealth, result.CurrentHealth) // Current health should equal max health initially
	assert.Equal(t, 0, result.RevivalCount)                 // Should start with 0 revivals
	assert.Equal(t, 3, result.MaxRevivals)                 // Default max revivals
	assert.NotZero(t, result.CreatedAt)
}

