package arena

import (
    "context"
    "errors"

    "gorm.io/gorm"
)

type sqlRepo struct{}

func getGorm() (*gorm.DB, error) {
    if !SQLDB.Enabled { return nil, errors.New("sql not enabled") }
    db, ok := SQLDB.DB.(*gorm.DB)
    if !ok || db == nil { return nil, errors.New("invalid sql handle") }
    return db, nil
}

func (r *sqlRepo) GetMatchByID(ctx context.Context, id string) (*ArenaMatch, error) {
    db, err := getGorm(); if err != nil { return nil, err }
    var m ArenaMatchSQL
    if tx := db.WithContext(ctx).First(&m, "id = ?", id); tx.Error != nil { return nil, tx.Error }
    // map to ArenaMatch
    am := &ArenaMatch{
        // ID isn’t available as ObjectID; leave zero, fields carry state
        Player1ID: m.Player1ID, Player1Name: m.Player1Name,
        Player1HP: m.Player1HP, Player1MaxHP: m.Player1MaxHP,
        Player1Attack: m.Player1Attack, Player1Defense: m.Player1Defense,
        Player2ID: m.Player2ID, Player2Name: m.Player2Name,
        Player2HP: m.Player2HP, Player2MaxHP: m.Player2MaxHP,
        Player2Attack: m.Player2Attack, Player2Defense: m.Player2Defense,
        CurrentTurn: m.CurrentTurn, MaxTurns: m.MaxTurns, CurrentAttacker: m.CurrentAttacker,
        Status: ArenaMatchStatus(m.Status), WinnerID: m.WinnerID, WinnerName: m.WinnerName,
        P1Below50Announced: m.P1Below50Announced, P2Below50Announced: m.P2Below50Announced,
        P1Below10Announced: m.P1Below10Announced, P2Below10Announced: m.P2Below10Announced,
    }
    return am, nil
}

func (r *sqlRepo) UpdateMatchFields(ctx context.Context, id string, fields map[string]interface{}) error {
    db, err := getGorm(); if err != nil { return err }
    return db.WithContext(ctx).Model(&ArenaMatchSQL{}).Where("id = ?", id).Updates(fields).Error
}


