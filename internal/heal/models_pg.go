package heal

import "time"

// HealingRecordSQL is the SQL model for healing records
type HealingRecordSQL struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	WarriorID    uint      `gorm:"index;not null"`
	WarriorName  string    `gorm:"size:255"`
	HealType     string    `gorm:"size:32;not null"`
	HealedAmount int       `gorm:"not null"`
	HPBefore     int       `gorm:"not null"`
	HPAfter      int       `gorm:"not null"`
	CoinsSpent   int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (HealingRecordSQL) TableName() string {
	return "healing_records"
}

