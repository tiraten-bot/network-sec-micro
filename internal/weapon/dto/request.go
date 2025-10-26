package dto

// CreateWeaponRequest represents a weapon creation request
type CreateWeaponRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"required,max=500"`
	Type        string `json:"type" binding:"required,oneof=common rare"`
	Damage      int    `json:"damage" binding:"required,min=1,max=1000"`
	Price       int    `json:"price" binding:"required,min=1"`
}

// BuyWeaponRequest represents a weapon purchase request
type BuyWeaponRequest struct {
	WeaponID string `json:"weapon_id" binding:"required"`
}

// GetWeaponsByTypeRequest represents a query request
type GetWeaponsByTypeRequest struct {
	Type string `form:"type" binding:"omitempty,oneof=common rare legendary"`
}
