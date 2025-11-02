package dto

import (
	"network-sec-micro/internal/battle"
	"time"
)

// ParticipantResponse represents a battle participant response
type ParticipantResponse struct {
	ID          string    `json:"id"`
	BattleID    string    `json:"battle_id"`
	ParticipantID string  `json:"participant_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Side        string    `json:"side"`
	HP          int       `json:"hp"`
	MaxHP       int       `json:"max_hp"`
	AttackPower int       `json:"attack_power"`
	Defense     int       `json:"defense"`
	IsAlive     bool      `json:"is_alive"`
	IsDefeated  bool      `json:"is_defeated"`
	DefeatedAt  *string   `json:"defeated_at,omitempty"`
	CreatedAt   string    `json:"created_at"`
}

// ToParticipantResponse converts a BattleParticipant to ParticipantResponse
func ToParticipantResponse(p *battle.BattleParticipant) *ParticipantResponse {
	resp := &ParticipantResponse{
		ID:            p.ID.Hex(),
		BattleID:      p.BattleID.Hex(),
		ParticipantID: p.ParticipantID,
		Name:          p.Name,
		Type:          string(p.Type),
		Side:          string(p.Side),
		HP:            p.HP,
		MaxHP:         p.MaxHP,
		AttackPower:   p.AttackPower,
		Defense:       p.Defense,
		IsAlive:       p.IsAlive,
		IsDefeated:    p.IsDefeated,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.DefeatedAt != nil {
		defeatedStr := p.DefeatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.DefeatedAt = &defeatedStr
	}

	return resp
}

// BattleResponse represents a battle response
type BattleResponse struct {
	ID                    string                `json:"id"`
	BattleType            string                `json:"battle_type"`
	LightSideName         string                `json:"light_side_name"`
	DarkSideName          string                `json:"dark_side_name"`
	CurrentTurn           int                   `json:"current_turn"`
	CurrentParticipantIndex int                 `json:"current_participant_index"`
	MaxTurns              int                   `json:"max_turns"`
	Status                string                `json:"status"`
	Result                string                `json:"result,omitempty"`
	WinnerSide            string                `json:"winner_side,omitempty"`
	CoinsEarned           map[string]int        `json:"coins_earned,omitempty"`
	ExperienceGained      map[string]int        `json:"experience_gained,omitempty"`
	CreatedBy             string                `json:"created_by"`
	LightParticipants     []ParticipantResponse `json:"light_participants"`
	DarkParticipants      []ParticipantResponse `json:"dark_participants"`
	StartedAt             *string               `json:"started_at,omitempty"`
	CompletedAt           *string               `json:"completed_at,omitempty"`
	CreatedAt             string                `json:"created_at"`
}

// ToBattleResponse converts a Battle model to BattleResponse
func ToBattleResponse(b *battle.Battle, lightParticipants []*battle.BattleParticipant, darkParticipants []*battle.BattleParticipant) *BattleResponse {
	resp := &BattleResponse{
		ID:                    b.ID.Hex(),
		BattleType:            string(b.BattleType),
		LightSideName:         b.LightSideName,
		DarkSideName:          b.DarkSideName,
		CurrentTurn:           b.CurrentTurn,
		CurrentParticipantIndex: b.CurrentParticipantIndex,
		MaxTurns:              b.MaxTurns,
		Status:                string(b.Status),
		CreatedBy:             b.CreatedBy,
		CreatedAt:             b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if b.Result != "" {
		resp.Result = string(b.Result)
		resp.WinnerSide = string(b.WinnerSide)
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

	// Convert participants
	if lightParticipants != nil {
		resp.LightParticipants = make([]ParticipantResponse, len(lightParticipants))
		for i, p := range lightParticipants {
			resp.LightParticipants[i] = *ToParticipantResponse(p)
		}
	}

	if darkParticipants != nil {
		resp.DarkParticipants = make([]ParticipantResponse, len(darkParticipants))
		for i, p := range darkParticipants {
			resp.DarkParticipants[i] = *ToParticipantResponse(p)
		}
	}

	return resp
}

// BattleTurnResponse represents a battle turn response
type BattleTurnResponse struct {
	ID            string `json:"id"`
	BattleID      string `json:"battle_id"`
	TurnNumber    int    `json:"turn_number"`
	AttackerID    string `json:"attacker_id"`
	AttackerName  string `json:"attacker_name"`
	AttackerType  string `json:"attacker_type"`
	AttackerSide  string `json:"attacker_side"`
	TargetID      string `json:"target_id"`
	TargetName    string `json:"target_name"`
	TargetType    string `json:"target_type"`
	TargetSide    string `json:"target_side"`
	DamageDealt   int    `json:"damage_dealt"`
	CriticalHit   bool   `json:"critical_hit"`
	TargetHPBefore int   `json:"target_hp_before"`
	TargetHPAfter  int   `json:"target_hp_after"`
	TargetDefeated bool  `json:"target_defeated"`
	CreatedAt     string `json:"created_at"`
}

// ToBattleTurnResponse converts a BattleTurn to BattleTurnResponse
func ToBattleTurnResponse(t *battle.BattleTurn) *BattleTurnResponse {
	return &BattleTurnResponse{
		ID:            t.ID.Hex(),
		BattleID:      t.BattleID.Hex(),
		TurnNumber:    t.TurnNumber,
		AttackerID:    t.AttackerID,
		AttackerName:  t.AttackerName,
		AttackerType:  string(t.AttackerType),
		AttackerSide:  string(t.AttackerSide),
		TargetID:      t.TargetID,
		TargetName:    t.TargetName,
		TargetType:    string(t.TargetType),
		TargetSide:    string(t.TargetSide),
		DamageDealt:   t.DamageDealt,
		CriticalHit:   t.CriticalHit,
		TargetHPBefore: t.TargetHPBefore,
		TargetHPAfter:  t.TargetHPAfter,
		TargetDefeated: t.TargetDefeated,
		CreatedAt:     t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
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
	WarriorID          uint    `json:"warrior_id"`
	TotalBattles       int     `json:"total_battles"`
	Wins               int     `json:"wins"`
	Losses             int     `json:"losses"`
	Draws              int     `json:"draws"`
	WinRate            float64 `json:"win_rate"`
	TeamBattles        int     `json:"team_battles"`
	TotalCoinsEarned   int     `json:"total_coins_earned"`
	TotalExperience    int     `json:"total_experience"`
}
