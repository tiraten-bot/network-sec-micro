package battle

import (
    "time"
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
	BattleResultLightVictory BattleResult = "light_victory" // Light side won
	BattleResultDarkVictory  BattleResult = "dark_victory"  // Dark side won
	BattleResultDraw         BattleResult = "draw"          // Draw
)

// BattleType represents the type of battle
type BattleType string

const (
	BattleTypeTeam BattleType = "team" // Team battle (light vs dark)
    // Legacy single battles
    BattleTypeEnemy  BattleType = "enemy"
    BattleTypeDragon BattleType = "dragon"
)

// TeamSide represents which side a participant is on
type TeamSide string

const (
	TeamSideLight TeamSide = "light" // Light side (good)
	TeamSideDark  TeamSide = "dark"   // Dark side (evil)
)

// ParticipantType represents the type of participant
type ParticipantType string

const (
	ParticipantTypeWarrior ParticipantType = "warrior" // Warrior (knight, archer, mage, light_emperor, light_king)
	ParticipantTypeEnemy    ParticipantType = "enemy"   // Enemy (goblin, pirate, orc, etc.)
	ParticipantTypeDragon   ParticipantType = "dragon"  // Dragon (fire, ice, lightning, shadow)
	ParticipantTypeDarkKing  ParticipantType = "dark_king"
	ParticipantTypeDarkEmperor ParticipantType = "dark_emperor"
)

// BattleParticipant represents a single participant in a battle
type BattleParticipant struct {
    ID            string             `json:"id"`
    BattleID      string             `json:"battle_id"`
	ParticipantID string             `bson:"participant_id" json:"participant_id"` // Warrior ID, Enemy ID, or Dragon ID
	Name          string             `bson:"name" json:"name"`
	Type          ParticipantType    `bson:"type" json:"type"`
	Side          TeamSide           `bson:"side" json:"side"` // light or dark
	
	// Stats
	HP            int                `bson:"hp" json:"hp"`
	MaxHP         int                `bson:"max_hp" json:"max_hp"`
	AttackPower   int                `bson:"attack_power" json:"attack_power"`
	Defense       int                `bson:"defense" json:"defense"`
	
	// Status
	IsAlive       bool               `bson:"is_alive" json:"is_alive"`
	IsDefeated   bool               `bson:"is_defeated" json:"is_defeated"`
	DefeatedAt   *time.Time         `bson:"defeated_at,omitempty" json:"defeated_at,omitempty"`
	
    CreatedAt    time.Time          `json:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (BattleParticipant) CollectionName() string {
	return "battle_participants"
}

// Battle represents a team-based battle
type Battle struct {
    ID            string             `json:"id"`
	BattleType    BattleType         `bson:"battle_type" json:"battle_type"`
	
	// Team information
	LightSideName string             `bson:"light_side_name" json:"light_side_name"` // e.g., "Light Alliance"
	DarkSideName  string             `bson:"dark_side_name" json:"dark_side_name"`   // e.g., "Dark Forces"
	
	// Turn information
	CurrentTurn   int                `bson:"current_turn" json:"current_turn"`
	CurrentParticipantIndex int     `bson:"current_participant_index" json:"current_participant_index"` // Index in turn order
	MaxTurns      int                `bson:"max_turns" json:"max_turns"`
	
	// Battle result
	Status        BattleStatus       `bson:"status" json:"status"`
	Result        BattleResult       `bson:"result,omitempty" json:"result,omitempty"`
	WinnerSide    TeamSide           `bson:"winner_side,omitempty" json:"winner_side,omitempty"`
	
	// Rewards (calculated after battle)
	CoinsEarned   map[string]int     `bson:"coins_earned,omitempty" json:"coins_earned,omitempty"` // participant_id -> coins
	ExperienceGained map[string]int  `bson:"experience_gained,omitempty" json:"experience_gained,omitempty"` // participant_id -> exp
	
	// Metadata
	CreatedBy     string             `bson:"created_by" json:"created_by"` // Creator username
    // Legacy single battle metadata (optional)
    WarriorID     uint               `json:"warrior_id,omitempty"`
    WarriorName   string             `json:"warrior_name,omitempty"`
    OpponentID    string             `json:"opponent_id,omitempty"`
    OpponentName  string             `json:"opponent_name,omitempty"`
    OpponentType  string             `json:"opponent_type,omitempty"`
    WarriorHP     int                `json:"warrior_hp,omitempty"`
    WarriorMaxHP  int                `json:"warrior_max_hp,omitempty"`
    OpponentHP    int                `json:"opponent_hp,omitempty"`
    OpponentMaxHP int                `json:"opponent_max_hp,omitempty"`
	
	// Timestamps
    StartedAt     *time.Time         `json:"started_at,omitempty"`
    CompletedAt   *time.Time         `json:"completed_at,omitempty"`
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`

    // Emperor wager (optional)
    WagerAmount   int                `json:"wager_amount"`
    LightEmperorID string            `json:"light_emperor_id"`
    DarkEmperorID  string            `json:"dark_emperor_id"`
    LightEmperorApproved bool        `json:"light_emperor_approved"`
    DarkEmperorApproved  bool        `json:"dark_emperor_approved"`
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
    ID            string             `json:"id"`
    BattleID      string             `json:"battle_id"`
	TurnNumber    int                `bson:"turn_number" json:"turn_number"`
	
	// Attacker info
	AttackerID    string             `bson:"attacker_id" json:"attacker_id"` // Participant ID
	AttackerName  string             `bson:"attacker_name" json:"attacker_name"`
	AttackerType  ParticipantType    `bson:"attacker_type" json:"attacker_type"`
	AttackerSide  TeamSide           `bson:"attacker_side" json:"attacker_side"`
	
	// Target info
	TargetID      string             `bson:"target_id" json:"target_id"` // Participant ID
	TargetName    string             `bson:"target_name" json:"target_name"`
	TargetType    ParticipantType    `bson:"target_type" json:"target_type"`
	TargetSide    TeamSide           `bson:"target_side" json:"target_side"`
	
	// Damage dealt
	DamageDealt   int                `bson:"damage_dealt" json:"damage_dealt"`
	CriticalHit   bool               `bson:"critical_hit" json:"critical_hit"`
	
	// HP before and after attack
	TargetHPBefore int               `bson:"target_hp_before" json:"target_hp_before"`
	TargetHPAfter int                `bson:"target_hp_after" json:"target_hp_after"`
	
	// Was target defeated in this attack?
	TargetDefeated bool              `bson:"target_defeated" json:"target_defeated"`
	
    CreatedAt     time.Time          `json:"created_at"`
}

// CollectionName returns the MongoDB collection name
func (BattleTurn) CollectionName() string {
	return "battle_turns"
}
