package battle

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BattleStatus represents the status of a battle
type BattleStatus string

const (
	BattleStatusPending   BattleStatus = "pending"   // Battle created but not started
	BattleStatusInProgress BattleStatus = "in_progress" // Battle is ongoing
	BattleStatusCompleted BattleStatus = "completed" // Battle finished
	BattleStatusCancelled BattleStatus = "cancelled" // Battle cancelled
)

// BattleResult represents the result of a battle
type BattleResult string

const (
	BattleResultVictory BattleResult = "victory" // Warrior won
	BattleResultDefeat  BattleResult = "defeat"  // Warrior lost
	BattleResultDraw     BattleResult = "draw"    // Draw (if applicable)
)

// BattleType represents the type of battle
type BattleType string

const (
	BattleTypeEnemy  BattleType = "enemy"  // Battle against enemy
	BattleTypeDragon BattleType = "dragon" // Battle against dragon
	BattleTypeArena  BattleType = "arena"  // PvP arena battle (future)
)

// Battle represents a battle record
type Battle struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BattleType    BattleType         `bson:"battle_type" json:"battle_type"`
	WarriorID     uint               `bson:"warrior_id" json:"warrior_id"`
	WarriorName   string             `bson:"warrior_name" json:"warrior_name"`
	OpponentID    string             `bson:"opponent_id" json:"opponent_id"` // Enemy or Dragon ID
	OpponentName  string             `bson:"opponent_name" json:"opponent_name"`
	OpponentType  string             `bson:"opponent_type" json:"opponent_type"`
	
	// Battle stats
	WarriorHP     int                `bson:"warrior_hp" json:"warrior_hp"`
	WarriorMaxHP  int                `bson:"warrior_max_hp" json:"warrior_max_hp"`
	OpponentHP    int                `bson:"opponent_hp" json:"opponent_hp"`
	OpponentMaxHP int                `bson:"opponent_max_hp" json:"opponent_max_hp"`
	
	// Turn information
	CurrentTurn   int                `bson:"current_turn" json:"current_turn"`
	MaxTurns      int                `bson:"max_turns" json:"max_turns"`
	
	// Battle result
	Status        BattleStatus       `bson:"status" json:"status"`
	Result        BattleResult       `bson:"result,omitempty" json:"result,omitempty"`
	WinnerID      string             `bson:"winner_id,omitempty" json:"winner_id,omitempty"`
	WinnerName    string             `bson:"winner_name,omitempty" json:"winner_name,omitempty"`
	
	// Rewards (if won)
	CoinsEarned   int                `bson:"coins_earned,omitempty" json:"coins_earned,omitempty"`
	ExperienceGained int            `bson:"experience_gained,omitempty" json:"experience_gained,omitempty"`
	
	// Timestamps
	StartedAt     *time.Time         `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt   *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (Battle) CollectionName() string {
	return "battles"
}

// IsActive checks if battle is still active
func (b *Battle) IsActive() bool {
	return b.Status == BattleStatusInProgress || b.Status == BattleStatusPending
}

// BattleTurn represents a single turn in a battle
type BattleTurn struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BattleID      primitive.ObjectID `bson:"battle_id" json:"battle_id"`
	TurnNumber    int                `bson:"turn_number" json:"turn_number"`
	
	// Attacker info
	AttackerID    string             `bson:"attacker_id" json:"attacker_id"`
	AttackerName  string             `bson:"attacker_name" json:"attacker_name"`
	AttackerType  string             `bson:"attacker_type" json:"attacker_type"` // "warrior" or "opponent"
	
	// Target info
	TargetID      string             `bson:"target_id" json:"target_id"`
	TargetName    string             `bson:"target_name" json:"target_name"`
	TargetType    string             `bson:"target_type" json:"target_type"`
	
	// Damage dealt
	DamageDealt   int                `bson:"damage_dealt" json:"damage_dealt"`
	CriticalHit   bool               `bson:"critical_hit" json:"critical_hit"`
	
	// HP after attack
	TargetHPAfter int                `bson:"target_hp_after" json:"target_hp_after"`
	
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// CollectionName returns the MongoDB collection name
func (BattleTurn) CollectionName() string {
	return "battle_turns"
}

