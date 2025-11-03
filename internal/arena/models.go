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

// ArenaMatch represents an active or completed arena match
type ArenaMatch struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	
	// Players
	Player1ID   uint   `bson:"player1_id" json:"player1_id"`
	Player1Name string `bson:"player1_name" json:"player1_name"`
	Player2ID   uint   `bson:"player2_id" json:"player2_id"`
	Player2Name string `bson:"player2_name" json:"player2_name"`
	
	// Battle info
	BattleID    string `bson:"battle_id" json:"battle_id"` // Battle service'deki battle ID
	
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

