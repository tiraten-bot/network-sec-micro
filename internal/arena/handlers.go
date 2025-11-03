package arena

import (
	"net/http"

	"network-sec-micro/internal/arena/dto"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

