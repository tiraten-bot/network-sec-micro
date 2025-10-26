package dto

// GetWarriorQuery represents a query to get a warrior by ID
type GetWarriorQuery struct {
	WarriorID uint
}

// GetWarriorsByRoleQuery represents a query to get warriors by role
type GetWarriorsByRoleQuery struct {
	Role string
}

// GetAllWarriorsQuery represents a query to get all warriors
type GetAllWarriorsQuery struct {
	Limit  int
	Offset int
}

// GetWarriorByUsernameQuery represents a query to get a warrior by username
type GetWarriorByUsernameQuery struct {
	Username string
}

// GetWarriorByEmailQuery represents a query to get a warrior by email
type GetWarriorByEmailQuery struct {
	Email string
}
