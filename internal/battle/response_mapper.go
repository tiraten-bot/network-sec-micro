package battle

import (
    "network-sec-micro/internal/battle/dto"
)

func ToParticipantResponse(p *BattleParticipant) *dto.ParticipantResponse {
    resp := &dto.ParticipantResponse{
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

func ToBattleResponse(b *Battle, lightParticipants []*BattleParticipant, darkParticipants []*BattleParticipant) *dto.BattleResponse {
    resp := &dto.BattleResponse{
        ID:                      b.ID.Hex(),
        BattleType:              string(b.BattleType),
        LightSideName:           b.LightSideName,
        DarkSideName:            b.DarkSideName,
        CurrentTurn:             b.CurrentTurn,
        CurrentParticipantIndex: b.CurrentParticipantIndex,
        MaxTurns:                b.MaxTurns,
        Status:                  string(b.Status),
        CreatedBy:               b.CreatedBy,
        CreatedAt:               b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
    if lightParticipants != nil {
        resp.LightParticipants = make([]dto.ParticipantResponse, len(lightParticipants))
        for i, p := range lightParticipants {
            resp.LightParticipants[i] = *ToParticipantResponse(p)
        }
    }
    if darkParticipants != nil {
        resp.DarkParticipants = make([]dto.ParticipantResponse, len(darkParticipants))
        for i, p := range darkParticipants {
            resp.DarkParticipants[i] = *ToParticipantResponse(p)
        }
    }
    return resp
}

func ToBattleTurnResponse(t *BattleTurn) *dto.BattleTurnResponse {
    return &dto.BattleTurnResponse{
        ID:             t.ID.Hex(),
        BattleID:       t.BattleID.Hex(),
        TurnNumber:     t.TurnNumber,
        AttackerID:     t.AttackerID,
        AttackerName:   t.AttackerName,
        AttackerType:   string(t.AttackerType),
        AttackerSide:   string(t.AttackerSide),
        TargetID:       t.TargetID,
        TargetName:     t.TargetName,
        TargetType:     string(t.TargetType),
        TargetSide:     string(t.TargetSide),
        DamageDealt:    t.DamageDealt,
        CriticalHit:    t.CriticalHit,
        TargetHPBefore: t.TargetHPBefore,
        TargetHPAfter:  t.TargetHPAfter,
        TargetDefeated: t.TargetDefeated,
        CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
    }
}


