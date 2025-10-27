package dto

// CreateWeaponCommand represents a command to create a weapon
type CreateWeaponCommand struct {
	Name        string
	Description string
	Type        string
	Damage      int
	Price       int
	CreatedBy   string
}

// BuyWeaponCommand represents a command to buy a weapon
type BuyWeaponCommand struct {
	WeaponID      string
	BuyerRole     string
	BuyerID       string // Username
	BuyerUsername string // Display name (same as BuyerID typically)
	BuyerUserID   uint   // Numeric ID from warrior service
}
