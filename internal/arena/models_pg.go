package arena

import "time"

// SQL models for gradual migration to PostgreSQL via GORM

// ArenaInvitationSQL mirrors ArenaInvitation for relational storage
type ArenaInvitationSQL struct {
    ID            uint   `gorm:"primaryKey;autoIncrement"`
    ChallengerID  uint   `gorm:"not null;index"`
    ChallengerName string `gorm:"size:255;not null"`
    OpponentID    uint   `gorm:"not null;index"`
    OpponentName  string `gorm:"size:255;not null"`
    Status        string `gorm:"size:32;not null;index"`
    ExpiresAt     time.Time
    RespondedAt   *time.Time
    BattleID      string `gorm:"size:64"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

func (ArenaInvitationSQL) TableName() string { return "arena_invitations" }

// ArenaMatchSQL mirrors ArenaMatch for relational storage
type ArenaMatchSQL struct {
    ID            uint   `gorm:"primaryKey;autoIncrement"`
    Player1ID     uint   `gorm:"index"`
    Player1Name   string `gorm:"size:255"`
    Player1HP     int
    Player1MaxHP  int
    Player1Attack int
    Player1Defense int
    Player2ID     uint   `gorm:"index"`
    Player2Name   string `gorm:"size:255"`
    Player2HP     int
    Player2MaxHP  int
    Player2Attack int
    Player2Defense int
    CurrentTurn   int
    MaxTurns      int
    CurrentAttacker uint
    Status        string `gorm:"size:32;index"`
    WinnerID      *uint
    WinnerName    string `gorm:"size:255"`
    StartedAt     *time.Time
    CompletedAt   *time.Time
    P1Below50Announced bool `gorm:"not null;default:false"`
    P2Below50Announced bool `gorm:"not null;default:false"`
    P1Below10Announced bool `gorm:"not null;default:false"`
    P2Below10Announced bool `gorm:"not null;default:false"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

func (ArenaMatchSQL) TableName() string { return "arena_matches" }


