package dto

import "network-sec-micro/internal/battle"

// BattleResponse represents a battle response
type BattleResponse struct {
	ID            string            `json:"id"`
	BattleType    string            `json:"battle_type"`
	WarriorID     uint             `json:"warrior_id"`
	WarriorName   string           `json:"warrior_name"`
	OpponentID    string           `json:"opponent_id"`
	OpponentName  string           `json:"opponent_name"`
	OpponentType  string           `json:"opponent_type"`
	WarriorHP     int              `json:"warrior_hp"`
	WarriorMaxHP  int              `json:"warrior_max_hp"`
	OpponentHP    int              `json:"opponent_hp"`
	OpponentMaxHP int              `json:"opponent_max_hp"`
	CurrentTurn   int              `json:"current_turn"`
	MaxTurns      int              `json:"max_turns"`
	Status        string           `json:"status"`
	Result        string           `json:"result,omitempty"`
	WinnerName    string           `json:"winner_name,omitempty"`
	CoinsEarned   int              `json:"coins_earned,omitempty"`
	ExperienceGained int          `json:"experience_gained,omitempty"`
	StartedAt     *string          `json:"started_at,omitempty"`
	CompletedAt   *string          `json:"completed_at,omitempty"`
	CreatedAt     string           `json:"created_at"`
}

// ToBattleResponse converts a Battle model to BattleResponse
func ToBattleResponse(b *battle.Battle) *BattleResponse {
	resp := &BattleResponse{
		ID:            b.ID.Hex(),
		BattleType:    string(b.BattleType),
		WarriorID:     b.WarriorID,
		WarriorName:   b.WarriorName,
		OpponentID:    b.OpponentID,
		OpponentName:  b.OpponentName,
		OpponentType:  b.OpponentType,
		WarriorHP:     b.WarriorHP,
		WarriorMaxHP:  b.WarriorMaxHP,
		OpponentHP:    b.OpponentHP,
		OpponentMaxHP: b.OpponentMaxHP,
		CurrentTurn:   b.CurrentTurn,
		MaxTurns:      b.MaxTurns,
		Status:        string(b.Status),
		CreatedAt:     b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if b.Result != "" {
		resp.Result = string(b.Result)
		resp.WinnerName = b.WinnerName
		resp.CoinsEarned = b.CoinsEarned
		resp.ExperienceGained = b.ExperienceGained
	}

	if b.StartedAt != nil {
		startedStr := b.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.StartedAt = &startedStr
	}

	if b.CompletedAt != nil {
		completedStr := b.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &completedStr
	}

	return resp
}

// BattleTurnResponse represents a battle turn response
type BattleTurnResponse struct {
	ID            string `json:"id"`
	BattleID      string `json:"battle_id"`
	TurnNumber    int    `json:"turn_number"`
	AttackerName  string `json:"attacker_name"`
	AttackerType  string `json:"attacker_type"`
	TargetName    string `json:"target_name"`
	DamageDealt   int    `json:"damage_dealt"`
	CriticalHit   bool   `json:"critical_hit"`
	TargetHPAfter int    `json:"target_hp_after"`
	CreatedAt     string `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// BattlesListResponse represents a list of battles
type BattlesListResponse struct {
	Battles []*BattleResponse `json:"battles"`
	Count   int               `json:"count"`
	Total   int               `json:"total"`
}

// BattleStatsResponse represents battle statistics
type BattleStatsResponse struct {
	WarriorID          uint   `json:"warrior_id"`
	TotalBattles       int    `json:"total_battles"`
	Wins               int    `json:"wins"`
	Losses             int    `json:"losses"`
	Draws              int    `json:"draws"`
	WinRate            float64 `json:"win_rate"`
	EnemyBattles       int    `json:"enemy_battles"`
	DragonBattles      int    `json:"dragon_battles"`
	TotalCoinsEarned   int    `json:"total_coins_earned"`
	TotalExperience    int    `json:"total_experience"`
}

