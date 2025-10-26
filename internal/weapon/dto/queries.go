package dto

// GetWeaponsQuery represents a query to get weapons
type GetWeaponsQuery struct {
	Type     string
	OwnedBy  string
	CreatedBy string
}

// GetWeaponByIDQuery represents a query to get a weapon by ID
type GetWeaponByIDQuery struct {
	WeaponID string
}
