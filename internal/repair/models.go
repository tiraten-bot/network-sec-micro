package repair

import "time"

type RepairOrderStatus string

const (
    RepairStatusPending   RepairOrderStatus = "pending"
    RepairStatusCompleted RepairOrderStatus = "completed"
    RepairStatusFailed    RepairOrderStatus = "failed"
)

// RepairOrder SQL model (GORM)
type RepairOrder struct {
    ID         uint              `gorm:"primaryKey;autoIncrement"`
    OwnerType  string            `gorm:"size:32;index;not null"`
    OwnerID    string            `gorm:"size:255;index;not null"`
    WeaponID   string            `gorm:"size:64;index"` // Optional, for weapon repairs
    ArmorID    string            `gorm:"size:64;index"` // Optional, for armor repairs
    ItemType   string            `gorm:"size:32;index;not null"` // "weapon" | "armor"
    Cost       int               `gorm:"not null"`
    Status     RepairOrderStatus `gorm:"size:32;index;not null"`
    CreatedAt  time.Time         `gorm:"not null"`
    CompletedAt *time.Time
}

func (RepairOrder) TableName() string { return "repair_orders" }


