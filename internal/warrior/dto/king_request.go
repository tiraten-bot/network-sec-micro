package dto

// CreateKingRequest represents a king creation request
type CreateKingRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=light_king dark_king"`
}

// CreateKingCommand represents a command to create a king
type CreateKingCommand struct {
	Username  string
	Email     string
	Password  string
	Role      string
	CreatedBy uint
}
