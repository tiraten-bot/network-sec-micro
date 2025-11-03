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
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;not null" json:"username"`
	Email       string    `gorm:"uniqueIndex;not null" json:"email"`
	Password    string    `gorm:"not null" json:"-"`
	Role        Role      `gorm:"type:varchar(20);not null" json:"role"`
	CoinBalance int       `gorm:"default:1000" json:"coin_balance"`     // Starting coin balance
	TotalPower  int       `gorm:"default:100" json:"total_power"`       // Total attack power
	WeaponCount int       `gorm:"default:0" json:"weapon_count"`        // Number of owned weapons
	CurrentHP   int       `gorm:"default:0" json:"current_hp"`          // Current HP (for healing)
	MaxHP       int       `gorm:"default:0" json:"max_hp"`              // Maximum HP (calculated from total_power)
	IsHealing   bool      `gorm:"default:false" json:"is_healing"`      // Is currently healing
	HealingUntil *time.Time `json:"healing_until,omitempty"`            // When healing completes
    // Title and achievement counters
    Title            string    `gorm:"type:varchar(50);default:''" json:"title"`
    EnemyKillCount   int       `gorm:"default:0" json:"enemy_kill_count"`
    DragonKillCount  int       `gorm:"default:0" json:"dragon_kill_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KilledMonster represents a monster killed by a warrior
type KilledMonster struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    WarriorID       uint      `gorm:"index;not null" json:"warrior_id"`
    MonsterKind     string    `gorm:"type:varchar(20);not null" json:"monster_kind"` // "dragon" or "enemy"
    EnemyType       string    `gorm:"type:varchar(30)" json:"enemy_type,omitempty"`
    DragonType      string    `gorm:"type:varchar(30)" json:"dragon_type,omitempty"`
    MonsterID       string    `gorm:"type:varchar(60);not null" json:"monster_id"`
    MonsterName     string    `gorm:"type:varchar(100);not null" json:"monster_name"`
    Level           int       `gorm:"default:0" json:"level"`
    MaxHealth       int       `gorm:"default:0" json:"max_health"`
    AttackPower     int       `gorm:"default:0" json:"attack_power"`
    Defense         int       `gorm:"default:0" json:"defense"`
    KilledAt        time.Time `json:"killed_at"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// TableName specifies the table name for Warrior
func (Warrior) TableName() string {
	return "warriors"
}

// IsLightSide checks if the warrior is on the light side
func (w *Warrior) IsLightSide() bool {
	return w.Role == RoleLightEmperor || w.Role == RoleLightKing || 
		   w.Role == RoleKnight || w.Role == RoleArcher || w.Role == RoleMage
}

// IsDarkSide checks if the warrior is on the dark side
func (w *Warrior) IsDarkSide() bool {
	return w.Role == RoleDarkEmperor || w.Role == RoleDarkKing
}

// CanCreateWarriors checks if the warrior can create new warriors
func (w *Warrior) CanCreateWarriors() bool {
	return w.Role == RoleLightEmperor || w.Role == RoleLightKing
}

// CanCreateKings checks if the warrior can create kings
func (w *Warrior) CanCreateKings() bool {
	return w.Role == RoleLightEmperor || w.Role == RoleDarkEmperor
}

// IsEmperor checks if the warrior is an emperor
func (w *Warrior) IsEmperor() bool {
	return w.Role == RoleLightEmperor || w.Role == RoleDarkEmperor
}

// IsKing checks if the warrior is a king
func (w *Warrior) IsKing() bool {
	return w.Role == RoleLightKing || w.Role == RoleDarkKing
}

// HasPermission checks if a warrior has permission for a specific resource
func (w *Warrior) HasPermission(resource string) bool {
	// Light emperor and light king have access to warrior resources
	if w.CanCreateWarriors() {
		return true
	}

	// Define role-based permissions (only for light side warriors)
	permissions := map[Role][]string{
		RoleKnight: {"weapons", "armor", "battles"},
		RoleArcher: {"weapons", "arrows", "scouting"},
		RoleMage:   {"spells", "potions", "library"},
	}

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
	// Dark side cannot access warrior endpoints
	if w.IsDarkSide() {
		return false
	}

	// Light emperor and light king have access to all warrior endpoints
	if w.CanCreateWarriors() {
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

// GetRoleType returns the role type category
func (w *Warrior) GetRoleType() string {
	if w.IsEmperor() {
		return "emperor"
	}
	if w.IsKing() {
		return "king"
	}
	return "warrior"
}
