package dto

// SendInvitationRequest represents a request to send an arena invitation
type SendInvitationRequest struct {
	OpponentName string `json:"opponent_name" binding:"required"`
}

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	InvitationID string `json:"invitation_id" binding:"required"`
}

// RejectInvitationRequest represents a request to reject an invitation
type RejectInvitationRequest struct {
	InvitationID string `json:"invitation_id" binding:"required"`
}

// CancelInvitationRequest represents a request to cancel an invitation
type CancelInvitationRequest struct {
	InvitationID string `json:"invitation_id" binding:"required"`
}
