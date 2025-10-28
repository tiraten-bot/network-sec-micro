package kafka

import "time"

type EnemyAttackEvent struct {
	Event
	EnemyID     string `json:"enemy_id"`
	EnemyType   string `json:"enemy_type"`
	EnemyName   string `json:"enemy_name"`
	WarriorID   uint   `json:"warrior_id"`
	WarriorName string `json:"warrior_name"`
	AttackType  string `json:"attack_type"`
	StolenValue int    `json:"stolen_value"`
	WeaponID    string `json:"weapon_id,omitempty"`
}

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

const TopicEnemyAttack = "enemy.attack"
