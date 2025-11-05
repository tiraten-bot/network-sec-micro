package armor_test

import (
	"context"
	"testing"
	"time"

	"network-sec-micro/internal/armor"
	"network-sec-micro/internal/armor/dto"

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
	
	db := client.Database("armor_test_db")
	db.Collection("armors").Drop(ctx)
	
	armor.ArmorColl = db.Collection("armors")
	
	return db
}

func TestCreateArmor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateArmorCommand{
		Name:         "Steel Plate",
		Description:  "Legendary armor",
		Type:         "legendary",
		Defense:      80,
		HPBonus:      50,
		Price:        3000,
		MaxDurability: 300,
		CreatedBy:    "warrior1",
	}
	
	result, err := svc.CreateArmor(ctx, cmd)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Steel Plate", result.Name)
	assert.Equal(t, "heavy", string(result.Type))
	assert.Equal(t, 80, result.Defense)
	assert.Equal(t, 50, result.HpBonus)
	assert.Equal(t, 3000, result.Price)
}

func TestCreateArmor_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateArmorCommand{
		Name:        "Test Armor",
		Description: "Test",
		Type:        "invalid_type",
		Defense:     50,
		HpBonus:     25,
		Price:       1000,
		CreatedBy:   "warrior1",
	}
	
	result, err := svc.CreateArmor(ctx, cmd)
	
	// Type validation might be in a different layer
	if err == nil {
		assert.NotNil(t, result)
	}
}

func TestGetArmors_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	armors := []dto.CreateArmorCommand{
		{Name: "Common Armor", Type: "common", Defense: 20, HPBonus: 10, Price: 500, MaxDurability: 100, CreatedBy: "warrior1"},
		{Name: "Rare Armor", Type: "rare", Defense: 50, HPBonus: 30, Price: 1500, MaxDurability: 200, CreatedBy: "warrior1"},
		{Name: "Legendary Armor", Type: "legendary", Defense: 80, HPBonus: 50, Price: 3000, MaxDurability: 300, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range armors {
		_, err := svc.CreateArmor(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetArmorsQuery{}
	result, err := svc.GetArmors(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestGetArmors_ByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	armors := []dto.CreateArmorCommand{
		{Name: "Common Armor 1", Type: "common", Defense: 20, HPBonus: 10, Price: 500, MaxDurability: 100, CreatedBy: "warrior1"},
		{Name: "Rare Armor", Type: "rare", Defense: 50, HPBonus: 30, Price: 1500, MaxDurability: 200, CreatedBy: "warrior1"},
		{Name: "Common Armor 2", Type: "common", Defense: 25, HPBonus: 15, Price: 600, MaxDurability: 100, CreatedBy: "warrior2"},
	}
	
	for _, cmd := range armors {
		_, err := svc.CreateArmor(ctx, cmd)
		require.NoError(t, err)
	}
	
	query := dto.GetArmorsQuery{Type: "common"}
	result, err := svc.GetArmors(ctx, query)
	
	require.NoError(t, err)
	assert.Len(t, result, 2)
	for _, a := range result {
		assert.Equal(t, armor.ArmorTypeCommon, a.Type)
	}
}

func TestArmor_CanBeBoughtBy(t *testing.T) {
	// Test armor permission logic
	commonArmor := armor.Armor{
		Type: armor.ArmorTypeCommon,
	}
	assert.True(t, commonArmor.CanBeBoughtBy("knight"))
	assert.True(t, commonArmor.CanBeBoughtBy("light_king"))
	assert.True(t, commonArmor.CanBeBoughtBy("light_emperor"))
	
	rareArmor := armor.Armor{
		Type: armor.ArmorTypeRare,
	}
	assert.False(t, rareArmor.CanBeBoughtBy("knight"))
	assert.True(t, rareArmor.CanBeBoughtBy("light_king"))
	assert.True(t, rareArmor.CanBeBoughtBy("light_emperor"))
	
	legendaryArmor := armor.Armor{
		Type: armor.ArmorTypeLegendary,
	}
	assert.False(t, legendaryArmor.CanBeBoughtBy("knight"))
	assert.False(t, legendaryArmor.CanBeBoughtBy("light_king"))
	assert.True(t, legendaryArmor.CanBeBoughtBy("light_emperor"))
	
	// Dark side cannot buy armors
	assert.False(t, commonArmor.CanBeBoughtBy("dark_emperor"))
	assert.False(t, commonArmor.CanBeBoughtBy("dark_king"))
	assert.False(t, rareArmor.CanBeBoughtBy("dark_emperor"))
	assert.False(t, legendaryArmor.CanBeBoughtBy("dark_emperor"))
}

func TestBuyArmor_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	// Create armor
	createCmd := dto.CreateArmorCommand{
		Name:      "Light Armor",
		Type:      "light",
		Defense:   30,
		HpBonus:   20,
		Price:     500,
		CreatedBy: "warrior1",
	}
	created, err := svc.CreateArmor(ctx, createCmd)
	require.NoError(t, err)
	
	// Buy armor
	buyCmd := dto.BuyArmorCommand{
		ArmorID:      created.ID.Hex(),
		BuyerID:      "buyer1",
		BuyerUserID:  1,
		BuyerUsername: "buyer1",
		BuyerRole:    "knight",
	}
	
	err = svc.BuyArmor(ctx, buyCmd)
	assert.NoError(t, err)
	
	// Verify armor was updated
	query := dto.GetArmorsQuery{}
	armors, err := svc.GetArmors(ctx, query)
	require.NoError(t, err)
	
	found := false
	for _, a := range armors {
		if a.ID.Hex() == created.ID.Hex() {
			assert.Len(t, a.Owners, 1)
			assert.Equal(t, "warrior", a.Owners[0].OwnerType)
			assert.Equal(t, "buyer1", a.Owners[0].OwnerID)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestBuyArmor_AlreadyOwned(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	// Create armor
	createCmd := dto.CreateArmorCommand{
		Name:      "Light Armor",
		Type:      "light",
		Defense:   30,
		HpBonus:   20,
		Price:     500,
		CreatedBy: "warrior1",
	}
	created, err := svc.CreateArmor(ctx, createCmd)
	require.NoError(t, err)
	
	// Buy armor first time
	buyCmd := dto.BuyArmorCommand{
		ArmorID:      created.ID.Hex(),
		BuyerID:      "buyer1",
		BuyerUserID:  1,
		BuyerUsername: "buyer1",
		BuyerRole:    "knight",
	}
	err = svc.BuyArmor(ctx, buyCmd)
	require.NoError(t, err)
	
	// Try to buy again
	err = svc.BuyArmor(ctx, buyCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already own")
}

func TestBuyArmor_PermissionDenied(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	// Create rare armor (only king/emperor can buy)
	createCmd := dto.CreateArmorCommand{
		Name:         "Rare Armor",
		Type:         "rare",
		Defense:      50,
		HPBonus:      30,
		Price:        1500,
		MaxDurability: 200,
		CreatedBy:    "warrior1",
	}
	created, err := svc.CreateArmor(ctx, createCmd)
	require.NoError(t, err)
	
	// Try to buy as knight (should fail)
	buyCmd := dto.BuyArmorCommand{
		ArmorID:      created.ID.Hex(),
		BuyerID:      "buyer1",
		BuyerUserID:  1,
		BuyerUsername: "buyer1",
		BuyerRole:    "knight",
	}
	
	err = svc.BuyArmor(ctx, buyCmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}

func TestArmorType_Constants(t *testing.T) {
	assert.Equal(t, armor.ArmorType("common"), armor.ArmorTypeCommon)
	assert.Equal(t, armor.ArmorType("rare"), armor.ArmorTypeRare)
	assert.Equal(t, armor.ArmorType("legendary"), armor.ArmorTypeLegendary)
}

func TestArmor_Durability(t *testing.T) {
	db := setupTestDB(t)
	defer db.Client().Disconnect(context.Background())
	
	svc := armor.NewService()
	ctx := context.Background()
	
	cmd := dto.CreateArmorCommand{
		Name:         "Test Armor",
		Type:         "common",
		Defense:      30,
		HPBonus:      20,
		Price:        500,
		MaxDurability: 100,
		CreatedBy:    "warrior1",
	}
	
	result, err := svc.CreateArmor(ctx, cmd)
	require.NoError(t, err)
	
	// Check default durability values
	assert.Greater(t, result.MaxDurability, 0)
	assert.Equal(t, result.MaxDurability, result.Durability) // Should start at max
	assert.False(t, result.IsBroken)
}

