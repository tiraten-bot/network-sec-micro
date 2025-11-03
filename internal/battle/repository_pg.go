package battle

import (
    "context"
    "errors"
    "fmt"

    "gorm.io/gorm"
)

type sqlRepo struct{}

func getGorm() (*gorm.DB, error) {
    if !SQLDB.Enabled { return nil, errors.New("sql not enabled") }
    db, ok := SQLDB.DB.(*gorm.DB)
    if !ok || db == nil { return nil, errors.New("invalid sql handle") }
    return db, nil
}

func (r *sqlRepo) GetBattleByID(ctx context.Context, id string) (*Battle, error) {
    db, err := getGorm(); if err != nil { return nil, err }
    var row BattleSQL
    if tx := db.WithContext(ctx).First(&row, "id = ?", id); tx.Error != nil { return nil, tx.Error }
    b := &Battle{
        ID: fmt.Sprintf("%d", row.ID),
        BattleType: BattleType(row.BattleType),
        LightSideName: row.LightSideName,
        DarkSideName: row.DarkSideName,
        CurrentTurn: row.CurrentTurn,
        CurrentParticipantIndex: row.CurrentParticipantIndex,
        MaxTurns: row.MaxTurns,
        Status: BattleStatus(row.Status),
        Result: BattleResult(row.Result),
        WinnerSide: TeamSide(row.WinnerSide),
        CreatedBy: row.CreatedBy,
        StartedAt: row.StartedAt,
        CompletedAt: row.CompletedAt,
        CreatedAt: row.CreatedAt,
        UpdatedAt: row.UpdatedAt,
    }
    return b, nil
}

func (r *sqlRepo) CreateBattle(ctx context.Context, b *Battle) (string, error) {
    db, err := getGorm(); if err != nil { return "", err }
    row := &BattleSQL{
        BattleType: string(b.BattleType),
        LightSideName: b.LightSideName,
        DarkSideName: b.DarkSideName,
        CurrentTurn: b.CurrentTurn,
        CurrentParticipantIndex: b.CurrentParticipantIndex,
        MaxTurns: b.MaxTurns,
        Status: string(b.Status),
        Result: string(b.Result),
        WinnerSide: string(b.WinnerSide),
        CreatedBy: b.CreatedBy,
        StartedAt: b.StartedAt,
        CompletedAt: b.CompletedAt,
        CreatedAt: b.CreatedAt,
        UpdatedAt: b.UpdatedAt,
    }
    if tx := db.WithContext(ctx).Create(row); tx.Error != nil { return "", tx.Error }
    return fmt.Sprintf("%d", row.ID), nil
}

func (r *sqlRepo) InsertParticipants(ctx context.Context, participants []*BattleParticipant) error {
    db, err := getGorm(); if err != nil { return err }
    rows := make([]*BattleParticipantSQL, 0, len(participants))
    for _, p := range participants {
        var battleID uint
        fmt.Sscanf(p.BattleID, "%d", &battleID)
        rows = append(rows, &BattleParticipantSQL{
            BattleID: battleID,
            ParticipantID: p.ParticipantID,
            Name: p.Name,
            Type: string(p.Type),
            Side: string(p.Side),
            HP: p.HP,
            MaxHP: p.MaxHP,
            AttackPower: p.AttackPower,
            Defense: p.Defense,
            IsAlive: p.IsAlive,
            IsDefeated: p.IsDefeated,
            DefeatedAt: p.DefeatedAt,
            CreatedAt: p.CreatedAt,
            UpdatedAt: p.UpdatedAt,
        })
    }
    return db.WithContext(ctx).Create(&rows).Error
}

func (r *sqlRepo) UpdateBattleFields(ctx context.Context, id string, fields map[string]interface{}) error {
    db, err := getGorm(); if err != nil { return err }
    return db.WithContext(ctx).Model(&BattleSQL{}).Where("id = ?", id).Updates(fields).Error
}

func (r *sqlRepo) GetParticipantByIDs(ctx context.Context, battleID string, participantID string) (*BattleParticipant, error) {
    db, err := getGorm(); if err != nil { return nil, err }
    var bid uint
    fmt.Sscanf(battleID, "%d", &bid)
    var row BattleParticipantSQL
    tx := db.WithContext(ctx).Where("battle_id = ? AND participant_id = ? AND is_alive = ?", bid, participantID, true).First(&row)
    if tx.Error != nil { return nil, tx.Error }
    p := &BattleParticipant{
        ID: fmt.Sprintf("%d", row.ID),
        BattleID: fmt.Sprintf("%d", row.BattleID),
        ParticipantID: row.ParticipantID,
        Name: row.Name,
        Type: ParticipantType(row.Type),
        Side: TeamSide(row.Side),
        HP: row.HP,
        MaxHP: row.MaxHP,
        AttackPower: row.AttackPower,
        Defense: row.Defense,
        IsAlive: row.IsAlive,
        IsDefeated: row.IsDefeated,
        DefeatedAt: row.DefeatedAt,
        CreatedAt: row.CreatedAt,
        UpdatedAt: row.UpdatedAt,
    }
    return p, nil
}

func (r *sqlRepo) UpdateParticipantByIDs(ctx context.Context, battleID string, participantID string, fields map[string]interface{}) error {
    db, err := getGorm(); if err != nil { return err }
    var bid uint
    fmt.Sscanf(battleID, "%d", &bid)
    return db.WithContext(ctx).Model(&BattleParticipantSQL{}).Where("battle_id = ? AND participant_id = ?", bid, participantID).Updates(fields).Error
}

func (r *sqlRepo) InsertTurn(ctx context.Context, turn *BattleTurn) error {
    db, err := getGorm(); if err != nil { return err }
    var bid uint
    fmt.Sscanf(turn.BattleID, "%d", &bid)
    row := &BattleTurnSQL{
        BattleID: bid,
        TurnNumber: turn.TurnNumber,
        AttackerID: turn.AttackerID,
        AttackerName: turn.AttackerName,
        AttackerType: string(turn.AttackerType),
        AttackerSide: string(turn.AttackerSide),
        TargetID: turn.TargetID,
        TargetName: turn.TargetName,
        TargetType: string(turn.TargetType),
        TargetSide: string(turn.TargetSide),
        DamageDealt: turn.DamageDealt,
        CriticalHit: turn.CriticalHit,
        TargetHPBefore: turn.TargetHPBefore,
        TargetHPAfter: turn.TargetHPAfter,
        TargetDefeated: turn.TargetDefeated,
        CreatedAt: turn.CreatedAt,
    }
    return db.WithContext(ctx).Create(row).Error
}

func (r *sqlRepo) FindParticipants(ctx context.Context, battleID string, sideFilter string) ([]*BattleParticipant, error) {
    db, err := getGorm(); if err != nil { return nil, err }
    var bid uint
    fmt.Sscanf(battleID, "%d", &bid)
    var rows []BattleParticipantSQL
    query := db.WithContext(ctx).Where("battle_id = ?", bid)
    if sideFilter != "all" && sideFilter != "" {
        query = query.Where("side = ?", sideFilter)
    }
    if err := query.Find(&rows).Error; err != nil { return nil, err }
    out := make([]*BattleParticipant, 0, len(rows))
    for _, rrow := range rows {
        rp := rrow
        out = append(out, &BattleParticipant{
            ID: fmt.Sprintf("%d", rp.ID),
            BattleID: fmt.Sprintf("%d", rp.BattleID),
            ParticipantID: rp.ParticipantID,
            Name: rp.Name,
            Type: ParticipantType(rp.Type),
            Side: TeamSide(rp.Side),
            HP: rp.HP,
            MaxHP: rp.MaxHP,
            AttackPower: rp.AttackPower,
            Defense: rp.Defense,
            IsAlive: rp.IsAlive,
            IsDefeated: rp.IsDefeated,
            DefeatedAt: rp.DefeatedAt,
            CreatedAt: rp.CreatedAt,
            UpdatedAt: rp.UpdatedAt,
        })
    }
    return out, nil
}

func (r *sqlRepo) CountAliveBySide(ctx context.Context, battleID string, side TeamSide) (int, error) {
    db, err := getGorm(); if err != nil { return 0, err }
    var bid uint
    fmt.Sscanf(battleID, "%d", &bid)
    var count int64
    if err := db.WithContext(ctx).Model(&BattleParticipantSQL{}).Where("battle_id = ? AND side = ? AND is_alive = ?", bid, string(side), true).Count(&count).Error; err != nil {
        return 0, err
    }
    return int(count), nil
}


