package arena

import (
    "time"
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
    ID            string                `json:"id"`
	
	// Challenger (sender)
    ChallengerID   uint   `json:"challenger_id"`
    ChallengerName string `json:"challenger_name"`
	
	// Opponent (receiver)
    OpponentID   uint   `json:"opponent_id"`
    OpponentName string `json:"opponent_name"`
	
	// Status
    Status     ArenaInvitationStatus `json:"status"`
    ExpiresAt  time.Time            `json:"expires_at"`
    RespondedAt *time.Time          `json:"responded_at,omitempty"`
	
	// Battle info (set when accepted)
    BattleID   string `json:"battle_id,omitempty"` // Battle service'den gelen battle ID
	
	// Timestamps
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
// legacy: kept for reference; no longer used with Postgres
func (ArenaInvitation) CollectionName() string { return "arena_invitations" }

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
    ID          string             `json:"id"`
	
	// Players
    Player1ID     uint   `json:"player1_id"`
    Player1Name   string `json:"player1_name"`
    Player1HP     int    `json:"player1_hp"`
    Player1MaxHP  int    `json:"player1_max_hp"`
    Player1Attack int    `json:"player1_attack"`
    Player1Defense int   `json:"player1_defense"`
	
    Player2ID     uint   `json:"player2_id"`
    Player2Name   string `json:"player2_name"`
    Player2HP     int    `json:"player2_hp"`
    Player2MaxHP  int    `json:"player2_max_hp"`
    Player2Attack int    `json:"player2_attack"`
    Player2Defense int   `json:"player2_defense"`
	
	// Battle progress
    CurrentTurn int    `json:"current_turn"`
    MaxTurns    int    `json:"max_turns"`
    CurrentAttacker uint `json:"current_attacker"` // 1 or 2, indicates whose turn it is

	// Spell windows (threshold announcements)
    P1Below50Announced bool `json:"p1_below50_announced"`
    P2Below50Announced bool `json:"p2_below50_announced"`
    P1Below10Announced bool `json:"p1_below10_announced"`
    P2Below10Announced bool `json:"p2_below10_announced"`
	
	// Result
    Status      ArenaMatchStatus `json:"status"`
    WinnerID    *uint            `json:"winner_id,omitempty"`
    WinnerName  string           `json:"winner_name,omitempty"`
	
	// Timestamps
    StartedAt   *time.Time `json:"started_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}

// CollectionName returns the MongoDB collection name
// legacy: kept for reference; no longer used with Postgres
func (ArenaMatch) CollectionName() string { return "arena_matches" }

