package heal_test

import (
	"testing"

	"network-sec-micro/internal/heal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHealPackageByType_FullHeal(t *testing.T) {
	packageInfo, err := heal.GetHealPackageByType(heal.HealTypeFull, "knight")
	
	require.NoError(t, err)
	assert.Equal(t, heal.HealTypeFull, packageInfo.Type)
	assert.Equal(t, 100, packageInfo.HealPercentage)
	assert.Equal(t, "warrior", packageInfo.RequiredRole)
}

func TestGetHealPackageByType_PartialHeal(t *testing.T) {
	packageInfo, err := heal.GetHealPackageByType(heal.HealTypePartial, "knight")
	
	require.NoError(t, err)
	assert.Equal(t, heal.HealTypePartial, packageInfo.Type)
	assert.Equal(t, 50, packageInfo.HealPercentage)
	assert.Equal(t, "warrior", packageInfo.RequiredRole)
}

func TestGetHealPackageByType_EmperorFullHeal(t *testing.T) {
	packageInfo, err := heal.GetHealPackageByType(heal.HealTypeEmperorFull, "light_emperor")
	
	require.NoError(t, err)
	assert.Equal(t, heal.HealTypeEmperorFull, packageInfo.Type)
	assert.Equal(t, 100, packageInfo.HealPercentage)
	assert.Equal(t, "emperor", packageInfo.RequiredRole)
}

func TestGetHealPackageByType_EmperorPartialHeal(t *testing.T) {
	packageInfo, err := heal.GetHealPackageByType(heal.HealTypeEmperorPartial, "dark_emperor")
	
	require.NoError(t, err)
	assert.Equal(t, heal.HealTypeEmperorPartial, packageInfo.Type)
	assert.Equal(t, 50, packageInfo.HealPercentage)
	assert.Equal(t, "emperor", packageInfo.RequiredRole)
}

func TestGetHealPackageByType_DragonHeal(t *testing.T) {
	packageInfo, err := heal.GetHealPackageByType(heal.HealTypeDragon, "dragon")
	
	require.NoError(t, err)
	assert.Equal(t, heal.HealTypeDragon, packageInfo.Type)
	assert.Equal(t, "dragon", packageInfo.RequiredRole)
}

func TestGetHealPackageByType_InvalidType(t *testing.T) {
	_, err := heal.GetHealPackageByType(heal.HealType("invalid"), "knight")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid heal type")
}

func TestGetHealPackageByType_RolePermission_Knight(t *testing.T) {
	// Knight can use warrior packages
	_, err := heal.GetHealPackageByType(heal.HealTypeFull, "knight")
	assert.NoError(t, err)
	
	_, err = heal.GetHealPackageByType(heal.HealTypePartial, "knight")
	assert.NoError(t, err)
	
	// Knight cannot use emperor packages
	_, err = heal.GetHealPackageByType(heal.HealTypeEmperorFull, "knight")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use")
	
	// Knight cannot use dragon packages
	_, err = heal.GetHealPackageByType(heal.HealTypeDragon, "knight")
	assert.Error(t, err)
}

func TestGetHealPackageByType_RolePermission_Emperor(t *testing.T) {
	// Light emperor can use emperor packages
	_, err := heal.GetHealPackageByType(heal.HealTypeEmperorFull, "light_emperor")
	assert.NoError(t, err)
	
	// Dark emperor can use emperor packages
	_, err = heal.GetHealPackageByType(heal.HealTypeEmperorFull, "dark_emperor")
	assert.NoError(t, err)
	
	// Emperor can also use warrior packages
	_, err = heal.GetHealPackageByType(heal.HealTypeFull, "light_emperor")
	assert.NoError(t, err)
}

func TestGetHealPackageByType_RolePermission_Dragon(t *testing.T) {
	// Dragon can use dragon packages
	_, err := heal.GetHealPackageByType(heal.HealTypeDragon, "dragon")
	assert.NoError(t, err)
	
	// Dragon cannot use emperor packages
	_, err = heal.GetHealPackageByType(heal.HealTypeEmperorFull, "dragon")
	assert.Error(t, err)
}

func TestHealPackage_Constants(t *testing.T) {
	// Test heal type constants
	assert.Equal(t, heal.HealType("full"), heal.HealTypeFull)
	assert.Equal(t, heal.HealType("partial"), heal.HealTypePartial)
	assert.Equal(t, heal.HealType("emperor_full"), heal.HealTypeEmperorFull)
	assert.Equal(t, heal.HealType("emperor_partial"), heal.HealTypeEmperorPartial)
	assert.Equal(t, heal.HealType("dragon"), heal.HealTypeDragon)
}

func TestHealPackage_Pricing(t *testing.T) {
	// Test that packages have prices
	full, err := heal.GetHealPackageByType(heal.HealTypeFull, "knight")
	require.NoError(t, err)
	assert.Equal(t, 100, full.Price)
	
	partial, err := heal.GetHealPackageByType(heal.HealTypePartial, "knight")
	require.NoError(t, err)
	assert.Equal(t, 50, partial.Price)
	
	emperorFull, err := heal.GetHealPackageByType(heal.HealTypeEmperorFull, "light_emperor")
	require.NoError(t, err)
	assert.Equal(t, 20, emperorFull.Price)
	
	emperorPartial, err := heal.GetHealPackageByType(heal.HealTypeEmperorPartial, "light_emperor")
	require.NoError(t, err)
	assert.Equal(t, 10, emperorPartial.Price)
	
	dragon, err := heal.GetHealPackageByType(heal.HealTypeDragon, "dragon")
	require.NoError(t, err)
	assert.Equal(t, 1000, dragon.Price)
}

func TestHealPackage_Duration(t *testing.T) {
	// Test healing durations
	full, err := heal.GetHealPackageByType(heal.HealTypeFull, "knight")
	require.NoError(t, err)
	assert.Equal(t, 300, full.Duration) // 5 minutes
	
	partial, err := heal.GetHealPackageByType(heal.HealTypePartial, "knight")
	require.NoError(t, err)
	assert.Equal(t, 180, partial.Duration) // 3 minutes
	
	emperorFull, err := heal.GetHealPackageByType(heal.HealTypeEmperorFull, "light_emperor")
	require.NoError(t, err)
	assert.Equal(t, 30, emperorFull.Duration) // 30 seconds
	
	emperorPartial, err := heal.GetHealPackageByType(heal.HealTypeEmperorPartial, "light_emperor")
	require.NoError(t, err)
	assert.Equal(t, 15, emperorPartial.Duration) // 15 seconds
	
	dragon, err := heal.GetHealPackageByType(heal.HealTypeDragon, "dragon")
	require.NoError(t, err)
	assert.Equal(t, 3600, dragon.Duration) // 1 hour
}

