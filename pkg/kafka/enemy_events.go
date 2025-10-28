package kafka

import "time"

// EnemyAttackEvent represents an enemy attack event
type EnemyAttackEvent struct {
	Event
	EnemyID     string `json:"enemy_id"`
	EnemyType   string `json:"enemy_type"`
	EnemyName   string `json:"enemy_name"`
	WarriorID   uint   `json:"warrior_id"`
	WarriorName string `json:"warrior_name"`
	AttackType  string `json:"attack_type"` // "coin_steal", "weapon_steal"
	StolenValue int    `json:"stolen_value"` // Amount stolen
	WeaponID    string `json:"weapon_id,omitempty"` // If weapon stolen
}

// NewGoblinCoinStealEvent creates a goblin coin steal event
func NewGoblinCoinStealEvent(enemyID, enemyName, warriorName string, warriorID, stolenCoins int) *EnemyAttackEvent {
	return &EnemyAttackEvent{
		Event: Event{
			EventType:     "enemy_attack",
			Timestamp:     time.Now(),
			SourceService: "enemy",
		},
		EnemyID:     enemyID,
		EnemyType:   "goblin",
		EnemyName:   enemyName,
		WarriorID:   uint(warriorID),
		WarriorName: warriorName,
		AttackType:  "coin_steal",
		StolenValue: stolenCoins,
	}
}

// NewPirateWeaponStealEvent creates a pirate weapon steal event
func NewPirateWeaponStealEvent(enemyID, enemyName, weaponID, warriorName string, warriorID int) *EnemyAttackEvent {
	return &EnemyAttackEvent{
		Event: Event{
			EventType:     "enemy_attack",
			Timestamp:     time.Now(),
			SourceService: "enemy",
		},
		EnemyID:     enemyID,
		EnemyType:   "pirate",
		EnemyName:   enemyName,
		WarriorID:   uint(warriorID),
		WarriorName: warriorName,
		AttackType:  "weapon_steal",
		WeaponID:    weaponID,
	}
}

// Topic names
const (
	TopicEnemyAttack = "enemy.attack"
)
