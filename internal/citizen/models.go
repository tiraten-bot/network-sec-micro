package citizen

import (
	"time"

	"gorm.io/gorm"
)

// Citizen represents a citizen user in the system
type Citizen struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	FirstName string    `gorm:"not null" json:"first_name"`
	LastName  string    `gorm:"not null" json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for Citizen
func (Citizen) TableName() string {
	return "citizens"
}

// BeforeCreate hook for Citizen
func (c *Citizen) BeforeCreate(tx *gorm.DB) error {
	return nil
}
