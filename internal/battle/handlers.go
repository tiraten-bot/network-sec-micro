package battle

import (
	"context"
	"net/http"
	"strconv"

	"network-sec-micro/internal/battle/dto"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Handler handles HTTP requests for battle service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler instance
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// GetCurrentUser extracts current user from JWT token
func GetCurrentUser(c *gin.Context) (*User, error) {
	username := c.GetString("username")
	if username == "" {
		return nil, gin.Error{
			Err:  nil,
			Type: gin.ErrorTypePublic,
			Meta: "username not found in token",
		}
	}

	userIDStr := c.GetString("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	return &User{
		Username: username,
		UserID:   uint(userID),
		Role:     c.GetString("role"),
	}, nil
}

// User represents authenticated user
type User struct {
	Username string
	UserID   uint
	Role     string
}

// StartBattle godoc
// @Summary Start a new battle
// @Description Start a battle against an enemy or dragon
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.StartBattleRequest true "Battle start data"
// @Success 201 {object} dto.BattleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /battles [post]
func (h *Handler) StartBattle(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req dto.StartBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// TODO: Fetch opponent info from enemy/dragon service
	// For now, we'll use placeholder values
	// In production, make HTTP/gRPC calls to get opponent details

	maxTurns := req.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 20
	}

	cmd := dto.StartBattleCommand{
		BattleType:    req.BattleType,
		WarriorID:     user.UserID,
		WarriorName:   user.Username,
		OpponentID:     req.OpponentID,
		OpponentType:  req.BattleType, // enemy or dragon
		OpponentName:  "Opponent",    // Should fetch from service
		MaxTurns:      maxTurns,
	}

	// Get opponent HP (placeholder - should fetch from enemy/dragon service)
	cmd.OpponentHP = 200 // Default

	battle, err := h.Service.StartBattle(cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "battle_start_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ToBattleResponse(battle))
}

// Attack godoc
// @Summary Perform an attack in battle
// @Description Warrior attacks opponent in an active battle
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AttackRequest true "Attack data"
// @Success 200 {object} map[string]interface{} "battle: BattleResponse, turn: BattleTurnResponse"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /battles/attack [post]
func (h *Handler) Attack(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req dto.AttackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	cmd := dto.AttackCommand{
		BattleID:    req.BattleID,
		WarriorID:   user.UserID,
		WarriorName: user.Username,
	}

	battle, turn, err := h.Service.Attack(cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "attack_failed",
			Message: err.Error(),
		})
		return
	}

	response := gin.H{
		"battle": dto.ToBattleResponse(battle),
	}

	if turn != nil {
		response["turn"] = &dto.BattleTurnResponse{
			ID:            turn.ID.Hex(),
			BattleID:      turn.BattleID.Hex(),
			TurnNumber:    turn.TurnNumber,
			AttackerName:  turn.AttackerName,
			AttackerType:  turn.AttackerType,
			TargetName:    turn.TargetName,
			DamageDealt:   turn.DamageDealt,
			CriticalHit:   turn.CriticalHit,
			TargetHPAfter: turn.TargetHPAfter,
			CreatedAt:     turn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetBattle godoc
// @Summary Get battle by ID
// @Description Get battle details by ID
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Battle ID"
// @Success 200 {object} dto.BattleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /battles/{id} [get]
func (h *Handler) GetBattle(c *gin.Context) {
	battleID := c.Param("id")
	if battleID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "battle ID is required",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(battleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid battle ID format",
		})
		return
	}

	query := dto.GetBattleQuery{
		BattleID: objectID,
	}

	battle, err := h.Service.GetBattle(query)
	if err != nil {
		if err.Error() == "battle not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Battle not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToBattleResponse(battle))
}

// GetMyBattles godoc
// @Summary Get my battles
// @Description Get list of battles for authenticated warrior
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (all, pending, in_progress, completed)"
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.BattlesListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /battles/my-battles [get]
func (h *Handler) GetMyBattles(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	status := c.DefaultQuery("status", "all")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	query := dto.GetBattlesByWarriorQuery{
		WarriorID: user.UserID,
		Status:    status,
		Limit:     limit,
		Offset:    offset,
	}

	battles, total, err := h.Service.GetBattlesByWarrior(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	responses := make([]*dto.BattleResponse, len(battles))
	for i, battle := range battles {
		responses[i] = dto.ToBattleResponse(&battle)
	}

	c.JSON(http.StatusOK, dto.BattlesListResponse{
		Battles: responses,
		Count:   len(responses),
		Total:   int(total),
	})
}

// GetBattleTurns godoc
// @Summary Get battle turns
// @Description Get turn history for a battle
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Battle ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{} "turns: []BattleTurnResponse, count: int"
// @Failure 400 {object} dto.ErrorResponse
// @Router /battles/{id}/turns [get]
func (h *Handler) GetBattleTurns(c *gin.Context) {
	battleID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(battleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid battle ID format",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	query := dto.GetBattleTurnsQuery{
		BattleID: objectID,
		Limit:    limit,
		Offset:   offset,
	}

	turns, err := h.Service.GetBattleTurns(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	responses := make([]*dto.BattleTurnResponse, len(turns))
	for i, turn := range turns {
		responses[i] = &dto.BattleTurnResponse{
			ID:            turn.ID.Hex(),
			BattleID:      turn.BattleID.Hex(),
			TurnNumber:    turn.TurnNumber,
			AttackerName:  turn.AttackerName,
			AttackerType:  turn.AttackerType,
			TargetName:    turn.TargetName,
			DamageDealt:   turn.DamageDealt,
			CriticalHit:   turn.CriticalHit,
			TargetHPAfter: turn.TargetHPAfter,
			CreatedAt:     turn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"turns": responses,
		"count": len(responses),
	})
}

// GetBattleStats godoc
// @Summary Get battle statistics
// @Description Get battle statistics for authenticated warrior
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by battle type (all, enemy, dragon)"
// @Success 200 {object} dto.BattleStatsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /battles/stats [get]
func (h *Handler) GetBattleStats(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	battleType := c.DefaultQuery("type", "all")

	query := dto.GetBattleStatsQuery{
		WarriorID:  user.UserID,
		BattleType: battleType,
	}

	stats, err := h.Service.GetBattleStats(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

