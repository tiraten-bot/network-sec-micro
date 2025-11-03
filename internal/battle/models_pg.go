package battle

import "time"

// SQL models for Battle service

type BattleSQL struct {
    ID                      uint   `gorm:"primaryKey;autoIncrement"`
    BattleType              string `gorm:"size:32;index"`
    LightSideName           string `gorm:"size:255"`
    DarkSideName            string `gorm:"size:255"`
    CurrentTurn             int
    CurrentParticipantIndex int
    MaxTurns                int
    Status                  string `gorm:"size:32;index"`
    Result                  string `gorm:"size:32"`
    WinnerSide              string `gorm:"size:16"`
    CreatedBy               string `gorm:"size:255"`
    StartedAt               *time.Time
    CompletedAt             *time.Time
    CreatedAt               time.Time
    UpdatedAt               time.Time
    WagerAmount             int
    LightEmperorID          string `gorm:"size:64"`
    DarkEmperorID           string `gorm:"size:64"`
    LightEmperorApproved    bool   `gorm:"not null;default:false"`
    DarkEmperorApproved     bool   `gorm:"not null;default:false"`
}

func (BattleSQL) TableName() string { return "battles" }

type BattleParticipantSQL struct {
    ID            uint   `gorm:"primaryKey;autoIncrement"`
    BattleID      uint   `gorm:"index;not null"`
    ParticipantID string `gorm:"size:64;index"`
    Name          string `gorm:"size:255"`
    Type          string `gorm:"size:32;index"`
    Side          string `gorm:"size:8;index"`
    HP            int
    MaxHP         int
    AttackPower   int
    Defense       int
    IsAlive       bool  `gorm:"not null;default:true"`
    IsDefeated    bool  `gorm:"not null;default:false"`
    DefeatedAt    *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

func (BattleParticipantSQL) TableName() string { return "battle_participants" }

type BattleTurnSQL struct {
    ID              uint   `gorm:"primaryKey;autoIncrement"`
    BattleID        uint   `gorm:"index;not null"`
    TurnNumber      int
    AttackerID      string `gorm:"size:64"`
    AttackerName    string `gorm:"size:255"`
    AttackerType    string `gorm:"size:32"`
    AttackerSide    string `gorm:"size:8"`
    TargetID        string `gorm:"size:64"`
    TargetName      string `gorm:"size:255"`
    TargetType      string `gorm:"size:32"`
    TargetSide      string `gorm:"size:8"`
    DamageDealt     int
    CriticalHit     bool
    TargetHPBefore  int
    TargetHPAfter   int
    TargetDefeated  bool
    CreatedAt       time.Time
}

func (BattleTurnSQL) TableName() string { return "battle_turns" }


