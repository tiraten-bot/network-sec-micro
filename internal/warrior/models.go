package warrior

import (
	"time"
)

// Role represents the role of a warrior
type Role string

const (
	// Warrior roles (can be created by light emperor/king)
	RoleKnight Role = "knight"
	RoleArcher Role = "archer"
	RoleMage   Role = "mage"
	
	// Light side leadership
	RoleLightEmperor Role = "light_emperor"
	RoleLightKing    Role = "light_king"
	
	// Dark side leadership
	RoleDarkEmperor Role = "dark_emperor"
	RoleDarkKing    Role = "dark_king"
)

// Warrior represents a warrior user in the system
type Warrior struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      Role      `gorm:"type:varchar(20);not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for Warrior
func (Warrior) TableName() string {
	return "warriors"
}

// HasPermission checks if a warrior has permission for a specific resource
func (w *Warrior) HasPermission(resource string) bool {
	// King has access to everything
	if w.Role == RoleKing {
		return true
	}

	// Define role-based permissions
	permissions := map[Role][]string{
		RoleKnight: {"weapons", "armor", "battles"},
		RoleArcher: {"weapons", "arrows", "scouting"},
		RoleMage:   {"spells", "potions", "library"},
	}

	// Check if the role has access to the resource
	if allowedResources, exists := permissions[w.Role]; exists {
		for _, allowed := range allowedResources {
			if allowed == resource {
				return true
			}
		}
	}

	return false
}

// CanAccessEndpoint checks if warrior can access a specific endpoint
func (w *Warrior) CanAccessEndpoint(endpoint string) bool {
	// King has access to all endpoints
	if w.Role == RoleKing {
		return true
	}

	// Define role-based endpoint access for warrior endpoints
	roleEndpoints := map[Role][]string{
		RoleKnight: []string{"/api/warriors/knights"},
		RoleArcher: []string{"/api/warriors/archers"},
		RoleMage:   []string{"/api/warriors/mages"},
	}

	if endpoints, exists := roleEndpoints[w.Role]; exists {
		for _, ep := range endpoints {
			if ep == endpoint {
				return true
			}
		}
	}

	return false
}

// IsKing checks if the warrior is a king (admin)
func (w *Warrior) IsKing() bool {
	return w.Role == RoleKing
}
