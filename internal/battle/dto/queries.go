package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

// GetBattleQuery represents a query to get a battle by ID
type GetBattleQuery struct {
	BattleID primitive.ObjectID `json:"battle_id"`
}

// GetBattlesByWarriorQuery represents a query to get battles for a warrior
type GetBattlesByWarriorQuery struct {
	WarriorID uint   `json:"warrior_id"` // 0 means all (for emperors/kings)
	Status    string `json:"status"`    // "all", "pending", "in_progress", "completed"
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// GetBattleTurnsQuery represents a query to get turns for a battle
type GetBattleTurnsQuery struct {
	BattleID primitive.ObjectID `json:"battle_id"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// GetBattleStatsQuery represents a query to get battle statistics
type GetBattleStatsQuery struct {
	WarriorID  uint   `json:"warrior_id"`
	BattleType string `json:"battle_type"` // "all", "team"
}

// GetBattleParticipantsQuery represents a query to get participants in a battle
type GetBattleParticipantsQuery struct {
	BattleID primitive.ObjectID `json:"battle_id"`
	Side     string             `json:"side"` // "all", "light", "dark"
}
