package arena

import (
	"net/http"

	"network-sec-micro/internal/arena/dto"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler handles HTTP requests for arena service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// SendInvitation godoc
// @Summary Send an arena invitation
// @Description Sends a 1v1 arena challenge invitation to another warrior. Anyone can challenge anyone (no role/rank restrictions).
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SendInvitationRequest true "Invitation data"
// @Success 201 {object} map[string]interface{} "invitation: ArenaInvitation"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations [post]
func (h *Handler) SendInvitation(c *gin.Context) {
	// TODO: Add auth middleware
	// user, err := GetCurrentUser(c)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
	// 		Error:   "unauthorized",
	// 		Message: err.Error(),
	// 	})
	// 	return
	// }

	var req dto.SendInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get user from context after auth middleware
	cmd := dto.SendInvitationCommand{
		ChallengerID:   1,          // TODO: Get from auth
		ChallengerName: "test_user", // TODO: Get from auth
		OpponentName:   req.OpponentName,
	}

	invitation, err := h.Service.SendInvitation(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invitation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"invitation": invitation,
		"message":    fmt.Sprintf("Invitation sent to %s", req.OpponentName),
	})
}

// AcceptInvitation godoc
// @Summary Accept an arena invitation
// @Description Accepts an arena invitation and starts the 1v1 match.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AcceptInvitationRequest true "Accept invitation data"
// @Success 200 {object} map[string]interface{} "match: ArenaMatch"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations/accept [post]
func (h *Handler) AcceptInvitation(c *gin.Context) {
	// TODO: Add auth middleware
	var req dto.AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get user from context
	cmd := dto.AcceptInvitationCommand{
		InvitationID: req.InvitationID,
		OpponentID:   2,        // TODO: Get from auth
		OpponentName: "user2", // TODO: Get from auth
	}

	match, err := h.Service.AcceptInvitation(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "accept_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match":   match,
		"message": "Invitation accepted! Arena match started.",
	})
}

// RejectInvitation godoc
// @Summary Reject an arena invitation
// @Description Rejects an arena invitation.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RejectInvitationRequest true "Reject invitation data"
// @Success 200 {object} map[string]interface{} "message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations/reject [post]
func (h *Handler) RejectInvitation(c *gin.Context) {
	var req dto.RejectInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get user from context
	cmd := dto.RejectInvitationCommand{
		InvitationID: req.InvitationID,
		OpponentID:   2, // TODO: Get from auth
	}

	if err := h.Service.RejectInvitation(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "reject_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation rejected",
	})
}

// CancelInvitation godoc
// @Summary Cancel an arena invitation
// @Description Cancels an arena invitation (only by the challenger).
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CancelInvitationRequest true "Cancel invitation data"
// @Success 200 {object} map[string]interface{} "message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations/cancel [post]
func (h *Handler) CancelInvitation(c *gin.Context) {
	var req dto.CancelInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get user from context
	cmd := dto.CancelInvitationCommand{
		InvitationID: req.InvitationID,
		ChallengerID: 1, // TODO: Get from auth
	}

	if err := h.Service.CancelInvitation(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "cancel_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation cancelled",
	})
}

// GetInvitation godoc
// @Summary Get an invitation by ID
// @Description Gets an arena invitation by ID.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invitation ID"
// @Success 200 {object} map[string]interface{} "invitation: ArenaInvitation"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations/{id} [get]
func (h *Handler) GetInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	if invitationID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invitation ID is required",
		})
		return
	}

	if _, err := primitive.ObjectIDFromHex(invitationID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid invitation ID format",
		})
		return
	}

	query := dto.GetInvitationQuery{
		InvitationID: invitationID,
	}

	invitation, err := h.Service.GetInvitation(c.Request.Context(), query)
	if err != nil {
		if err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Invitation not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitation": invitation,
	})
}

// GetMyInvitations godoc
// @Summary Get user's invitations
// @Description Gets all invitations (sent or received) for the current user.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, accepted, rejected, expired, cancelled)"
// @Success 200 {object} map[string]interface{} "invitations: []ArenaInvitation"
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/invitations/my [get]
func (h *Handler) GetMyInvitations(c *gin.Context) {
	// TODO: Get user from context
	status := c.Query("status")

	query := dto.GetMyInvitationsQuery{
		UserID: 1, // TODO: Get from auth
		Status: status,
	}

	invitations, err := h.Service.GetMyInvitations(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"count":       len(invitations),
	})
}

// GetMyMatches godoc
// @Summary Get user's arena matches
// @Description Gets all arena matches for the current user.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, in_progress, completed, cancelled)"
// @Success 200 {object} map[string]interface{} "matches: []ArenaMatch"
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/matches/my [get]
func (h *Handler) GetMyMatches(c *gin.Context) {
	// TODO: Get user from context
	status := c.Query("status")

	query := dto.GetMyMatchesQuery{
		UserID: 1, // TODO: Get from auth
		Status: status,
	}

	matches, err := h.Service.GetMyMatches(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"count":   len(matches),
	})
}

// AttackInArena godoc
// @Summary Perform an attack in arena match
// @Description Performs an attack in an active arena match. Players take turns attacking.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AttackInArenaRequest true "Attack data"
// @Success 200 {object} map[string]interface{} "match: ArenaMatch"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/arena/attack [post]
func (h *Handler) AttackInArena(c *gin.Context) {
	// TODO: Add auth middleware
	var req dto.AttackInArenaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	matchID, err := primitive.ObjectIDFromHex(req.MatchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid match ID format",
		})
		return
	}

	// TODO: Get user from context
	cmd := dto.AttackInArenaCommand{
		MatchID:    req.MatchID,
		AttackerID: 1, // TODO: Get from auth
	}

	match, err := h.Service.PerformAttack(c.Request.Context(), matchID, cmd.AttackerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "attack_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match":   match,
		"message": "Attack performed successfully",
	})
}

// ApplyArenaSpell godoc
// @Summary Apply an arenaspell effect to a match
// @Description Applies 1v1 spell effects (buff/debuff) directly to match stats
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ApplyArenaSpellRequest true "Spell apply data"
// @Success 200 {object} map[string]interface{} "match: ArenaMatch"
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/arena/spells/apply [post]
func (h *Handler) ApplyArenaSpell(c *gin.Context) {
    var req dto.ApplyArenaSpellRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
        return
    }

    matchID, err := primitive.ObjectIDFromHex(req.MatchID)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "invalid match ID format"})
        return
    }

    // Get caster from JWT
    user, err := GetCurrentUser(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()})
        return
    }

    // Enforce spell window at handler level as well (fast-fail)
    var matchSnapshot ArenaMatch
    if err := MatchColl.FindOne(c.Request.Context(), bson.M{"_id": matchID}).Decode(&matchSnapshot); err == nil {
        allow50 := func(hp, max int) bool { return max > 0 && (hp*100 <= max*50) }
        allow10 := func(hp, max int) bool { return max > 0 && (hp*100 <= max*10) }
        isCrisis := req.SpellType == "light_crisis" || req.SpellType == "dark_crisis"
        ok := false
        if isCrisis {
            ok = allow10(matchSnapshot.Player1HP, matchSnapshot.Player1MaxHP) || allow10(matchSnapshot.Player2HP, matchSnapshot.Player2MaxHP)
        } else {
            ok = allow50(matchSnapshot.Player1HP, matchSnapshot.Player1MaxHP) || allow50(matchSnapshot.Player2HP, matchSnapshot.Player2MaxHP)
        }
        if !ok {
            c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "spell window not open for this spell"})
            return
        }
    }

    // 1) Call arenaspell gRPC for RBAC ve state
    _, err = CastArenaSpellViaGRPC(c.Request.Context(), req.MatchID, req.SpellType, user.UserID, user.Username, user.Role)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "arenaspell_failed", Message: err.Error()})
        return
    }

    // 2) Apply actual effect to match (stats)
    match, err := h.Service.ApplySpellEffect(c.Request.Context(), matchID, user.UserID, req.SpellType)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "apply_failed", Message: err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "match":   match,
        "message": "Spell applied successfully",
    })
}

// GetMatch godoc
// @Summary Get an arena match by ID
// @Description Gets an arena match by ID.
// @Tags arena
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Match ID"
// @Success 200 {object} map[string]interface{} "match: ArenaMatch"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/arena/matches/{id} [get]
func (h *Handler) GetMatch(c *gin.Context) {
	matchID := c.Param("id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "match ID is required",
		})
		return
	}

	if _, err := primitive.ObjectIDFromHex(matchID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid match ID format",
		})
		return
	}

	// Get match from database
	var match ArenaMatch
	err := MatchColl.FindOne(c.Request.Context(), bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Match not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match": match,
	})
}

