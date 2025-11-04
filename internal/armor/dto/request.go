package dto

// CreateArmorRequest represents an armor creation request
type CreateArmorRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100"`
	Description  string `json:"description" binding:"required,max=500"`
	Type         string `json:"type" binding:"required,oneof=common rare"`
	Defense      int    `json:"defense" binding:"required,min=1,max=1000"`
	HPBonus      int    `json:"hp_bonus" binding:"required,min=0,max=2000"`
	Price        int    `json:"price" binding:"required,min=1"`
	MaxDurability int   `json:"max_durability" binding:"required,min=1,max=2000"`
}

// BuyArmorRequest represents an armor purchase request
type BuyArmorRequest struct {
	ArmorID   string `json:"armor_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"omitempty,oneof=warrior enemy dragon"` // Optional, defaults to warrior
}

// GetArmorsByTypeRequest represents a query request
type GetArmorsByTypeRequest struct {
	Type string `form:"type" binding:"omitempty,oneof=common rare legendary"`
}

