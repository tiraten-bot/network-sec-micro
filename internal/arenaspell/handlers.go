package arenaspell

import (
    "fmt"
    "net/http"

    "network-sec-micro/internal/arenaspell/dto"

    "github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for arenaspell service
type Handler struct {
    Service *Service
}

func NewHandler(s *Service) *Handler { return &Handler{Service: s} }

// CastSpell godoc
// @Summary Cast a spell in an arena match (1v1)
// @Description Applies a 1v1 arenaspell. Bufflar caster'a, debuff rakibe uygulanır.
// @Tags arenaspell
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CastArenaSpellRequest true "Spell casting data"
// @Success 200 {object} map[string]interface{} "success: bool, affected_count: int, message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Router /arenaspells/cast [post]
func (h *Handler) CastSpell(c *gin.Context) {
    var req dto.CastArenaSpellRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
        return
    }

    // TODO: Auth middleware ile gerçek kullanıcıyı al
    casterID := uint(1)
    casterUsername := "temp_user"

    cmd := dto.CastArenaSpellCommand{
        MatchID:        req.MatchID,
        SpellType:      req.SpellType,
        CasterUserID:   casterID,
        CasterUsername: casterUsername,
        CasterRole:     req.CasterRole,
    }

    affected, err := h.Service.CastSpell(c.Request.Context(), cmd)
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "spell_cast_failed", Message: err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success":        true,
        "affected_count": affected,
        "message":        fmt.Sprintf("Spell %s cast successfully!", req.SpellType),
    })
}


