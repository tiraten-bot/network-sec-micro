package dto

// GetInvitationQuery represents a query to get an invitation
type GetInvitationQuery struct {
	InvitationID string `json:"invitation_id"`
}

// GetMyInvitationsQuery represents a query to get user's invitations
type GetMyInvitationsQuery struct {
	UserID   uint   `json:"user_id"`
	Status   string `json:"status,omitempty"` // pending, accepted, rejected, expired, cancelled
}

// GetMyMatchesQuery represents a query to get user's matches
type GetMyMatchesQuery struct {
	UserID uint   `json:"user_id"`
	Status string `json:"status,omitempty"` // pending, in_progress, completed, cancelled
}

