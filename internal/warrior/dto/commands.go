package dto

// CreateWarriorCommand represents a command to create a warrior
type CreateWarriorCommand struct {
	Username string
	Email    string
	Password string
	Role     string
	CreatedBy uint
}

// UpdateWarriorCommand represents a command to update a warrior
type UpdateWarriorCommand struct {
	WarriorID uint
	Email     *string
	Role      *string
	UpdatedBy uint
}

// DeleteWarriorCommand represents a command to delete a warrior
type DeleteWarriorCommand struct {
	WarriorID uint
	DeletedBy uint
}

// ChangePasswordCommand represents a command to change password
type ChangePasswordCommand struct {
	WarriorID  uint
	OldPassword string
	NewPassword string
	ChangedBy  uint
}
