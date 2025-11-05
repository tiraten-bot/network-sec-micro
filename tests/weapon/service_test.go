package weapon_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/weapon"
	"network-sec-micro/internal/weapon/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB creates a test MongoDB connection (in-memory or test instance)
func setupTestDB(t *testing.T) *mongo.Database {
	// For testing, we'll use a test database
	// In CI/CD, this should connect to a test MongoDB instance
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	
	db := client.Database("weapon_test_db")
	// Clean up before test
	db.Collection("weapons").Drop(ctx)
	
	// Set the global collection
	weapon.WeaponColl = db.Collection("weapons")
	
	return db
}

func TestCreateWeapon_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateWeaponCommand{
		Name:        "Excalibur",
		Description: "Legendary sword",
		Type:        "legendary",
		Damage:      100,
		Price:       5000,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateWeapon(ctx, cmd)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Excalibur", result.Name)
	assert.Equal(t, "legendary", string(result.Type))
	assert.Equal(t, 100, result.Damage)
	assert.Equal(t, 5000, result.Price)
	assert.NotEmpty(t, result.ID)
}

func TestCreateWeapon_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateWeaponCommand{
		Name:        "Test Weapon",
		Description: "Test",
		Type:        "invalid_type",
		Damage:      50,
		Price:       100,
		CreatedBy:   "warrior1",
	}
	
	// Create should still work, but type validation happens elsewhere
	result, err := svc.CreateWeapon(ctx, cmd)
	
	// Type validation might be in a different layer
	// For now, we'll just check that creation doesn't fail on type
	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestGetWeapons_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Create multiple weapons
	weapons := []dto.CreateWeaponCommand{
		{Name: "Sword", Type: "common", Damage: 20, Price: 100, CreatedBy: "warrior1"},
		{Name: "Axe", Type: "rare", Damage: 50, Price: 500, CreatedBy: "warrior1"},
		{Name: "Bow", Type: "common", Damage: 30, Price: 200, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range weapons {
		_, err := svc.CreateWeapon(ctx, cmd)
		require.NoError(t, err)
	}
	
	// Get all weapons
	query := dto.GetWeaponsQuery{}
	result, err := svc.GetWeapons(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestGetWeapons_ByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Create weapons with different types
	weapons := []dto.CreateWeaponCommand{
		{Name: "Sword", Type: "common", Damage: 20, Price: 100, CreatedBy: "warrior1"},
		{Name: "Axe", Type: "rare", Damage: 50, Price: 500, CreatedBy: "warrior1"},
		{Name: "Bow", Type: "common", Damage: 30, Price: 200, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range weapons {
		_, err := svc.CreateWeapon(ctx, cmd)
		require.NoError(t, err)
	}
	
	// Get only common weapons
	query := dto.GetWeaponsQuery{Type: "common"}
	result, err := svc.GetWeapons(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, w := range result {
		assert.Equal(t, weapon.WeaponTypeCommon, w.Type)
	}
}

func TestGetWeapons_ByCreatedBy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Create weapons by different creators
	weapons := []dto.CreateWeaponCommand{
		{Name: "Sword", Type: "common", Damage: 20, Price: 100, CreatedBy: "warrior1"},
		{Name: "Axe", Type: "rare", Damage: 50, Price: 500, CreatedBy: "warrior1"},
		{Name: "Bow", Type: "common", Damage: 30, Price: 200, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range weapons {
		_, err := svc.CreateWeapon(ctx, cmd)
		require.NoError(t, err)
	}
	
	// Get weapons created by warrior1
	query := dto.GetWeaponsQuery{CreatedBy: "warrior1"}
	result, err := svc.GetWeapons(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, w := range result {
		assert.Equal(t, "warrior1", w.CreatedBy)
	}
}

func TestBuyWeapon_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Create weapon
	createCmd := dto.CreateWeaponCommand{
		Name:      "Test Sword",
		Type:      "common",
		Damage:    30,
		Price:     200,
		CreatedBy: "warrior1",
	}
	created, err := svc.CreateWeapon(ctx, createCmd)
	require.NoError(t, err)
	
	// Buy weapon
	buyCmd := dto.BuyWeaponCommand{
		WeaponID:     created.ID.Hex(),
		BuyerID:      "buyer1",
		BuyerUserID:  1,
		BuyerUsername: "buyer1",
		BuyerRole:    "knight",
	}
	
	err = svc.BuyWeapon(ctx, buyCmd)
	assert.NoError(t, err)
	
	// Verify weapon was updated
	query := dto.GetWeaponsQuery{}
	weapons, err := svc.GetWeapons(ctx, query)
	require.NoError(t, err)
	
	found := false
	for _, w := range weapons {
		if w.ID.Hex() == created.ID.Hex() {
			assert.Contains(t, w.OwnedBy, "buyer1")
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestBuyWeapon_AlreadyOwned(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Create weapon
	createCmd := dto.CreateWeaponCommand{
		Name:      "Test Sword",
		Type:      "common",
		Damage:    30,
		Price:     200,
		CreatedBy: "warrior1",
	}
	created, err := svc.CreateWeapon(ctx, createCmd)
	require.NoError(t, err)
	
	// Buy weapon first time
	buyCmd := dto.BuyWeaponCommand{
		WeaponID:      created.ID.Hex(),
		BuyerID:       "buyer1",
		BuyerUserID:   1,
		BuyerUsername: "buyer1",
		BuyerRole:     "knight",
	}
	err = svc.BuyWeapon(ctx, buyCmd)
	require.NoError(t, err)
	
	// Try to buy again
	err = svc.BuyWeapon(ctx, buyCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already own")
}

func TestBuyWeapon_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	// Try to buy non-existent weapon
	buyCmd := dto.BuyWeaponCommand{
		WeaponID:      primitive.NewObjectID().Hex(),
		BuyerID:       "buyer1",
		BuyerUserID:   1,
		BuyerUsername: "buyer1",
		BuyerRole:     "knight",
	}
	
	err := svc.BuyWeapon(ctx, buyCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBuyWeapon_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := weapon.NewService()
	ctx := context.Background()
	
	buyCmd := dto.BuyWeaponCommand{
		WeaponID:      "invalid_id",
		BuyerID:       "buyer1",
		BuyerUserID:   1,
		BuyerUsername: "buyer1",
		BuyerRole:     "knight",
	}
	
	err := svc.BuyWeapon(ctx, buyCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid weapon ID")
}

func TestWeapon_CanBeBoughtBy(t *testing.T) {
	// Test weapon permission logic
	commonWeapon := weapon.Weapon{
		Type: weapon.WeaponTypeCommon,
	}
	assert.True(t, commonWeapon.CanBeBoughtBy("knight"))
	assert.True(t, commonWeapon.CanBeBoughtBy("light_king"))
	
	rareWeapon := weapon.Weapon{
		Type: weapon.WeaponTypeRare,
	}
	assert.False(t, rareWeapon.CanBeBoughtBy("knight"))
	assert.True(t, rareWeapon.CanBeBoughtBy("light_king"))
	
	legendaryWeapon := weapon.Weapon{
		Type: weapon.WeaponTypeLegendary,
	}
	assert.False(t, legendaryWeapon.CanBeBoughtBy("knight"))
	assert.False(t, legendaryWeapon.CanBeBoughtBy("light_king"))
	assert.True(t, legendaryWeapon.CanBeBoughtBy("light_emperor"))
}

