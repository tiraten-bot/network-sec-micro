package dto

// CastArenaSpellRequest is the HTTP payload to cast a spell in an arena match
type CastArenaSpellRequest struct {
    MatchID       string `json:"match_id" binding:"required"`
    SpellType     string `json:"spell_type" binding:"required"`
}

// CastArenaSpellCommand is the command used by service layer
type CastArenaSpellCommand struct {
    MatchID         string
    SpellType       string
    CasterUserID    uint
    CasterUsername  string
}

// ErrorResponse is a generic error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}


