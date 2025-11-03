package dto

// SendInvitationCommand represents a command to send an arena invitation
type SendInvitationCommand struct {
	ChallengerID   uint   `json:"challenger_id"`
	ChallengerName string `json:"challenger_name"`
	OpponentName   string `json:"opponent_name"` // Opponent username
}

// AcceptInvitationCommand represents a command to accept an arena invitation
type AcceptInvitationCommand struct {
	InvitationID string `json:"invitation_id"`
	OpponentID   uint   `json:"opponent_id"` // Accepting user's ID
	OpponentName string `json:"opponent_name"`
}

// RejectInvitationCommand represents a command to reject an arena invitation
type RejectInvitationCommand struct {
	InvitationID string `json:"invitation_id"`
	OpponentID   uint   `json:"opponent_id"`
}

// CancelInvitationCommand represents a command to cancel an arena invitation
type CancelInvitationCommand struct {
	InvitationID string `json:"invitation_id"`
	ChallengerID uint   `json:"challenger_id"`
}

