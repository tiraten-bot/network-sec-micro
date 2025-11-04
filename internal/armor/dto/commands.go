package dto

// CreateArmorCommand represents a command to create an armor
type CreateArmorCommand struct {
	Name         string
	Description  string
	Type         string
	Defense      int
	HPBonus      int
	Price        int
	MaxDurability int
	CreatedBy    string
}

// BuyArmorCommand represents a command to buy an armor
type BuyArmorCommand struct {
	ArmorID      string
	BuyerRole    string
	BuyerID      string // Username or entity ID
	BuyerUsername string // Display name
	BuyerUserID  uint   // Numeric ID (for warrior)
	OwnerType    string // "warrior" | "enemy" | "dragon"
}

