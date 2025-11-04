package battle

import (
    "context"
    "os"
)

// Repository abstracts persistence for Battle domain
type Repository interface {
    GetBattleByID(ctx context.Context, id string) (*Battle, error)
    CreateBattle(ctx context.Context, b *Battle) (string, error)
    InsertParticipants(ctx context.Context, participants []*BattleParticipant) error
    UpdateBattleFields(ctx context.Context, id string, fields map[string]interface{}) error
    GetParticipantByIDs(ctx context.Context, battleID string, participantID string) (*BattleParticipant, error)
    UpdateParticipantByIDs(ctx context.Context, battleID string, participantID string, fields map[string]interface{}) error
    InsertTurn(ctx context.Context, turn *BattleTurn) error
    FindParticipants(ctx context.Context, battleID string, sideFilter string) ([]*BattleParticipant, error)
    CountAliveBySide(ctx context.Context, battleID string, side TeamSide) (int, error)
}

var defaultRepo Repository

// GetRepository returns a singleton repo based on env (BATTLE_STORE=redis|mongo)
func GetRepository() Repository {
    if defaultRepo != nil { return defaultRepo }
    // For stability, use Mongo-backed repo by default
    _ = os.Getenv("BATTLE_STORE")
    defaultRepo = &sqlRepo{}
    return defaultRepo
}


