package arena

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ArenaInvitationStatus represents the status of an arena invitation
type ArenaInvitationStatus string

const (
	InvitationStatusPending   ArenaInvitationStatus = "pending"   // Invitation sent, awaiting response
	InvitationStatusAccepted  ArenaInvitationStatus = "accepted"  // Invitation accepted
	InvitationStatusRejected  ArenaInvitationStatus = "rejected"  // Invitation rejected
	InvitationStatusExpired   ArenaInvitationStatus = "expired"   // Invitation expired
	InvitationStatusCancelled ArenaInvitationStatus = "cancelled" // Invitation cancelled by sender
)

// ArenaMatchStatus represents the status of an arena match
type ArenaMatchStatus string

const (
	MatchStatusPending   ArenaMatchStatus = "pending"   // Match created, waiting to start
	MatchStatusInProgress ArenaMatchStatus = "in_progress" // Match is ongoing
	MatchStatusCompleted ArenaMatchStatus = "completed" // Match finished
	MatchStatusCancelled ArenaMatchStatus = "cancelled" // Match cancelled
)

// ArenaInvitation represents a 1v1 arena challenge invitation
type ArenaInvitation struct {
	ID            primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	
	// Challenger (sender)
	ChallengerID   uint   `bson:"challenger_id" json:"challenger_id"`
	ChallengerName string `bson:"challenger_name" json:"challenger_name"`
	
	// Opponent (receiver)
	OpponentID   uint   `bson:"opponent_id" json:"opponent_id"`
	OpponentName string `bson:"opponent_name" json:"opponent_name"`
	
	// Status
	Status     ArenaInvitationStatus `bson:"status" json:"status"`
	ExpiresAt  time.Time            `bson:"expires_at" json:"expires_at"`
	RespondedAt *time.Time          `bson:"responded_at,omitempty" json:"responded_at,omitempty"`
	
	// Battle info (set when accepted)
	BattleID   string `bson:"battle_id,omitempty" json:"battle_id,omitempty"` // Battle service'den gelen battle ID
	
	// Timestamps
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (ArenaInvitation) CollectionName() string {
	return "arena_invitations"
}

// IsExpired checks if the invitation has expired
func (a *ArenaInvitation) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// CanBeAccepted checks if the invitation can be accepted
func (a *ArenaInvitation) CanBeAccepted() bool {
	return a.Status == InvitationStatusPending && !a.IsExpired()
}

// ArenaMatch represents an active or completed arena match (1v1)
type ArenaMatch struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Players
	Player1ID     uint   `bson:"player1_id" json:"player1_id"`
	Player1Name   string `bson:"player1_name" json:"player1_name"`
	Player1HP     int    `bson:"player1_hp" json:"player1_hp"`
	Player1MaxHP  int    `bson:"player1_max_hp" json:"player1_max_hp"`
	Player1Attack int    `bson:"player1_attack" json:"player1_attack"`
	Player1Defense int   `bson:"player1_defense" json:"player1_defense"`
	
	Player2ID     uint   `bson:"player2_id" json:"player2_id"`
	Player2Name   string `bson:"player2_name" json:"player2_name"`
	Player2HP     int    `bson:"player2_hp" json:"player2_hp"`
	Player2MaxHP  int    `bson:"player2_max_hp" json:"player2_max_hp"`
	Player2Attack int    `bson:"player2_attack" json:"player2_attack"`
	Player2Defense int   `bson:"player2_defense" json:"player2_defense"`
	
	// Battle progress
	CurrentTurn int    `bson:"current_turn" json:"current_turn"`
	MaxTurns    int    `bson:"max_turns" json:"max_turns"`
	CurrentAttacker uint `bson:"current_attacker" json:"current_attacker"` // 1 or 2, indicates whose turn it is

	// Spell windows (threshold announcements)
	P1Below50Announced bool `bson:"p1_below50_announced" json:"p1_below50_announced"`
	P2Below50Announced bool `bson:"p2_below50_announced" json:"p2_below50_announced"`
	P1Below10Announced bool `bson:"p1_below10_announced" json:"p1_below10_announced"`
	P2Below10Announced bool `bson:"p2_below10_announced" json:"p2_below10_announced"`
	
	// Result
	Status      ArenaMatchStatus `bson:"status" json:"status"`
	WinnerID    *uint            `bson:"winner_id,omitempty" json:"winner_id,omitempty"`
	WinnerName  string           `bson:"winner_name,omitempty" json:"winner_name,omitempty"`
	
	// Timestamps
	StartedAt   *time.Time `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
func (ArenaMatch) CollectionName() string {
	return "arena_matches"
}

