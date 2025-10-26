package warrior

import (
	"time"

	"gorm.io/gorm"
)

// Role represents the role of a warrior
type Role string

const (
	RoleKnight Role = "knight" // Knight warrior role
	RoleArcher Role = "archer" // Archer warrior role
	RoleMage   Role = "mage"   // Mage warrior role
	RoleKing   Role = "king"   // King role (admin - all access)
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
		RoleKnight:最低白羊{"/api/warriors/knights"},
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
