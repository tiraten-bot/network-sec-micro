package kafka

import "time"

// ArenaInvitationSentEvent represents when an arena invitation is sent
type ArenaInvitationSentEvent struct {
	Event
	InvitationID    string `json:"invitation_id"`
	ChallengerID    uint   `json:"challenger_id"`
	ChallengerName  string `json:"challenger_name"`
	OpponentID      uint   `json:"opponent_id"`
	OpponentName    string `json:"opponent_name"`
	ExpiresAt       string `json:"expires_at"`
}

// NewArenaInvitationSentEvent creates a new arena invitation sent event
func NewArenaInvitationSentEvent(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string, expiresAtStr string) *ArenaInvitationSentEvent {
	return &ArenaInvitationSentEvent{
		Event: Event{
			EventType:     "arena_invitation_sent",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		InvitationID:   invitationID,
		ChallengerID:   challengerID,
		ChallengerName: challengerName,
		OpponentID:     opponentID,
		OpponentName:   opponentName,
		ExpiresAt:      expiresAtStr,
	}
}

// ArenaInvitationAcceptedEvent represents when an arena invitation is accepted
type ArenaInvitationAcceptedEvent struct {
	Event
	InvitationID    string `json:"invitation_id"`
	ChallengerID    uint   `json:"challenger_id"`
	ChallengerName  string `json:"challenger_name"`
	OpponentID      uint   `json:"opponent_id"`
	OpponentName    string `json:"opponent_name"`
	BattleID        string `json:"battle_id"`
}

// NewArenaInvitationAcceptedEvent creates a new arena invitation accepted event
func NewArenaInvitationAcceptedEvent(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string, battleID string) *ArenaInvitationAcceptedEvent {
	return &ArenaInvitationAcceptedEvent{
		Event: Event{
			EventType:     "arena_invitation_accepted",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		InvitationID:   invitationID,
		ChallengerID: challengerID,
		ChallengerName: challengerName,
		OpponentID:   opponentID,
		OpponentName: opponentName,
		BattleID:     battleID,
	}
}

// ArenaInvitationRejectedEvent represents when an arena invitation is rejected
type ArenaInvitationRejectedEvent struct {
	Event
	InvitationID    string `json:"invitation_id"`
	ChallengerID    uint   `json:"challenger_id"`
	ChallengerName  string `json:"challenger_name"`
	OpponentID      uint   `json:"opponent_id"`
	OpponentName    string `json:"opponent_name"`
}

// NewArenaInvitationRejectedEvent creates a new arena invitation rejected event
func NewArenaInvitationRejectedEvent(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string) *ArenaInvitationRejectedEvent {
	return &ArenaInvitationRejectedEvent{
		Event: Event{
			EventType:     "arena_invitation_rejected",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		InvitationID:   invitationID,
		ChallengerID:   challengerID,
		ChallengerName: challengerName,
		OpponentID:     opponentID,
		OpponentName:   opponentName,
	}
}

// ArenaInvitationExpiredEvent represents when an arena invitation expires
type ArenaInvitationExpiredEvent struct {
	Event
	InvitationID    string `json:"invitation_id"`
	ChallengerID    uint   `json:"challenger_id"`
	ChallengerName  string `json:"challenger_name"`
	OpponentID      uint   `json:"opponent_id"`
	OpponentName    string `json:"opponent_name"`
}

// NewArenaInvitationExpiredEvent creates a new arena invitation expired event
func NewArenaInvitationExpiredEvent(invitationID string, challengerID uint, challengerName string, opponentID uint, opponentName string) *ArenaInvitationExpiredEvent {
	return &ArenaInvitationExpiredEvent{
		Event: Event{
			EventType:     "arena_invitation_expired",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		InvitationID:   invitationID,
		ChallengerID:   challengerID,
		ChallengerName: challengerName,
		OpponentID:     opponentID,
		OpponentName:   opponentName,
	}
}

// ArenaMatchStartedEvent represents when an arena match starts
type ArenaMatchStartedEvent struct {
	Event
	MatchID      string `json:"match_id"`
	Player1ID    uint   `json:"player1_id"`
	Player1Name  string `json:"player1_name"`
	Player2ID    uint   `json:"player2_id"`
	Player2Name  string `json:"player2_name"`
	BattleID     string `json:"battle_id"`
}

// NewArenaMatchStartedEvent creates a new arena match started event
func NewArenaMatchStartedEvent(matchID string, player1ID uint, player1Name string, player2ID uint, player2Name string, battleID string) *ArenaMatchStartedEvent {
	return &ArenaMatchStartedEvent{
		Event: Event{
			EventType:     "arena_match_started",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		MatchID:     matchID,
		Player1ID:   player1ID,
		Player1Name: player1Name,
		Player2ID:   player2ID,
		Player2Name: player2Name,
		BattleID:    battleID,
	}
}

// ArenaMatchCompletedEvent represents when an arena match completes
type ArenaMatchCompletedEvent struct {
	Event
	MatchID      string `json:"match_id"`
	Player1ID    uint   `json:"player1_id"`
	Player1Name  string `json:"player1_name"`
	Player2ID    uint   `json:"player2_id"`
	Player2Name  string `json:"player2_name"`
	WinnerID     *uint  `json:"winner_id,omitempty"`
	WinnerName   string `json:"winner_name,omitempty"`
	BattleID     string `json:"battle_id"`
}

// NewArenaMatchCompletedEvent creates a new arena match completed event
func NewArenaMatchCompletedEvent(matchID string, player1ID uint, player1Name string, player2ID uint, player2Name string, winnerID *uint, winnerName string, battleID string) *ArenaMatchCompletedEvent {
	return &ArenaMatchCompletedEvent{
		Event: Event{
			EventType:     "arena_match_completed",
			Timestamp:     time.Now(),
			SourceService: "arena",
		},
		MatchID:     matchID,
		Player1ID:   player1ID,
		Player1Name: player1Name,
		Player2ID:   player2ID,
		Player2Name: player2Name,
		WinnerID:    winnerID,
		WinnerName:  winnerName,
		BattleID:    battleID,
	}
}

// Topic names for arena events
const (
	TopicArenaInvitationSent    = "arena.invitation.sent"
	TopicArenaInvitationAccepted = "arena.invitation.accepted"
	TopicArenaInvitationRejected = "arena.invitation.rejected"
	TopicArenaInvitationExpired = "arena.invitation.expired"
	TopicArenaMatchStarted      = "arena.match.started"
	TopicArenaMatchCompleted    = "arena.match.completed"
)

