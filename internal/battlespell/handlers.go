package battlespell

import (
	"fmt"
	"net/http"

	"network-sec-micro/internal/battlespell/dto"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for battlespell service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// CastSpell godoc
// @Summary Cast a spell in battle
// @Description Casts a spell in an ongoing battle. Only light_king and dark_king can cast spells.
// @Tags battlespell
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CastSpellRequest true "Spell casting data"
// @Success 200 {object} map[string]interface{} "success: bool, affected_count: int, message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /spells/cast [post]
func (h *Handler) CastSpell(c *gin.Context) {
	// TODO: Add auth middleware
	// user, err := GetCurrentUser(c)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
	// 		Error:   "unauthorized",
	// 		Message: err.Error(),
	// 	})
	// 	return
	// }

	// Verify user is a king
	// if user.Role != "light_king" && user.Role != "dark_king" {
	// 	c.JSON(http.StatusForbidden, dto.ErrorResponse{
	// 		Error:   "forbidden",
	// 		Message: "only light_king and dark_king can cast spells",
	// 	})
	// 	c.Abort()
	// 	return
	// }

	var req dto.CastSpellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get user from context after auth middleware
	cmd := dto.CastSpellCommand{
		BattleID:            req.BattleID,
		SpellType:           req.SpellType,
		CasterUsername:      "temp_user", // TODO: Get from auth
		CasterUserID:        "temp_id",    // TODO: Get from auth
		CasterRole:          "light_king", // TODO: Get from auth
		TargetDragonID:      req.TargetDragonID,
		TargetDarkEmperorID: req.TargetDarkEmperorID,
	}

	affectedCount, err := h.Service.CastSpell(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "spell_cast_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"affected_count": affectedCount,
		"message":        fmt.Sprintf("Spell %s cast successfully! %d participants affected.", req.SpellType, affectedCount),
	})
}

