package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

// GetBattleQuery represents a query to get a battle by ID
type GetBattleQuery struct {
	BattleID primitive.ObjectID
}

// GetBattlesByWarriorQuery represents a query to get battles for a warrior
type GetBattlesByWarriorQuery struct {
	WarriorID uint
	Status    string // "all", "pending", "in_progress", "completed"
	Limit     int
	Offset    int
}

// GetBattleTurnsQuery represents a query to get turns for a battle
type GetBattleTurnsQuery struct {
	BattleID primitive.ObjectID
	Limit    int
	Offset   int
}

// GetBattleStatsQuery represents a query to get battle statistics
type GetBattleStatsQuery struct {
	WarriorID uint
	BattleType string // "all", "enemy", "dragon"
}

