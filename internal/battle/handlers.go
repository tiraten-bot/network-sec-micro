package battle

import (
	"context"
	"fmt"
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
// @Summary Start a new team battle
// @Description Start a team battle between Light and Dark sides. Light side can have warriors (knight, archer, mage, light_emperor, light_king). Dark side can have enemies, dragons, dark_king, dark_emperor.
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.StartBattleRequest true "Team battle start data"
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

	// Validate battle authorization: only emperors can start directly, kings need approvals
	ctx := c.Request.Context()
	if err := ValidateBattleAuthorization(ctx, user.Role, user.UserID, req.KingApprovals); err != nil {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "authorization_failed",
			Message: err.Error(),
		})
		return
	}

	maxTurns := req.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 100 // Default for team battles
	}

	cmd := dto.StartBattleCommand{
		LightSideName:      req.LightSideName,
		DarkSideName:       req.DarkSideName,
		LightParticipants:  req.LightParticipants,
		DarkParticipants:   req.DarkParticipants,
		MaxTurns:           maxTurns,
		CreatedBy:          user.Username,
	}

	battle, participants, err := h.Service.StartBattle(cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "battle_start_failed",
			Message: err.Error(),
		})
		return
	}

	// Separate participants by side
	var lightParts, darkParts []*BattleParticipant
	for _, p := range participants {
		if p.Side == TeamSideLight {
			lightParts = append(lightParts, p)
		} else {
			darkParts = append(darkParts, p)
		}
	}

	c.JSON(http.StatusCreated, dto.ToBattleResponse(battle, lightParts, darkParts))
}

// Attack godoc
// @Summary Perform an attack in team battle
// @Description A participant attacks another participant in an active team battle. Attacker and target must be on different sides.
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
		BattleID:     req.BattleID,
		AttackerID:   req.AttackerID,
		TargetID:     req.TargetID,
		AttackerName: user.Username, // For validation
		TargetName:   "",            // Will be fetched from participant
	}

	battle, turn, err := h.Service.Attack(cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "attack_failed",
			Message: err.Error(),
		})
		return
	}

	// Get participants for response
	battleID, _ := primitive.ObjectIDFromHex(req.BattleID)
	lightParts, _ := h.Service.GetBattleParticipants(c.Request.Context(), battleID, "light")
	darkParts, _ := h.Service.GetBattleParticipants(c.Request.Context(), battleID, "dark")

	response := gin.H{
		"battle": dto.ToBattleResponse(battle, lightParts, darkParts),
	}

	if turn != nil {
		response["turn"] = dto.ToBattleTurnResponse(turn)
	}

	c.JSON(http.StatusOK, response)
}

// GetBattle godoc
// @Summary Get battle by ID
// @Description Get battle details by ID. RBAC: Emperors see all, Kings see faction battles, Warriors see only their own.
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Battle ID"
// @Success 200 {object} dto.BattleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
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

	battle, lightParts, darkParts, err := h.Service.GetBattle(query)
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

	// RBAC check: Check if user is a participant or has admin access
	user, _ := GetCurrentUser(c)
	userIDStr := fmt.Sprintf("%d", user.UserID)
	hasAccess := false
	
	// Check if user is a participant
	for _, p := range lightParts {
		if p.ParticipantID == userIDStr {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		for _, p := range darkParts {
			if p.ParticipantID == userIDStr {
				hasAccess = true
				break
			}
		}
	}

	// Admin/emperor access
	if !hasAccess && !CheckBattleAccess(c, 0) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view this battle",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToBattleResponse(battle, lightParts, darkParts))
}

// GetMyBattles godoc
// @Summary Get battles
// @Description Get list of battles. Emperors see all battles. Kings see battles in their faction. Warriors see only their own battles.
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (all, pending, in_progress, completed)"
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Param warrior_id query int false "Warrior ID filter (emperors/kings only)"
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
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	// Apply RBAC: Emperors see all, Kings see faction, Warriors see only their own
	if err := GetBattlesWithRBAC(c, &query); err != nil {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: err.Error(),
		})
		return
	}

	// If emperor/king wants to filter by specific warrior
	if warriorIDStr := c.Query("warrior_id"); warriorIDStr != "" {
		canViewAll, _ := c.Get("can_view_all_battles")
		if canViewAll != nil && canViewAll.(bool) {
			if warriorIDFilter, err := strconv.ParseUint(warriorIDStr, 10, 32); err == nil {
				query.WarriorID = uint(warriorIDFilter)
			}
		}
	}

	battles, total, err := h.Service.GetBattlesByWarrior(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	// Get participants for each battle
	responses := make([]*dto.BattleResponse, len(battles))
	for i, battle := range battles {
		lightParts, _ := h.Service.GetBattleParticipants(c.Request.Context(), battle.ID, "light")
		darkParts, _ := h.Service.GetBattleParticipants(c.Request.Context(), battle.ID, "dark")
		responses[i] = dto.ToBattleResponse(&battle, lightParts, darkParts)
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
		responses[i] = dto.ToBattleTurnResponse(&turn)
	}

	c.JSON(http.StatusOK, gin.H{
		"turns": responses,
		"count": len(responses),
	})
}

// GetBattleStats godoc
// @Summary Get battle statistics
// @Description Get battle statistics. Warriors see only their own stats. Emperors/Kings can view any warrior's stats via warrior_id query param.
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by battle type (all, enemy, dragon)"
// @Param warrior_id query int false "Warrior ID (emperors/kings only - to view other warrior's stats)"
// @Success 200 {object} dto.BattleStatsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
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
	warriorID := user.UserID

	// RBAC: Emperors/Kings can view any warrior's stats
	if warriorIDStr := c.Query("warrior_id"); warriorIDStr != "" {
		canViewAll, _ := c.Get("can_view_all_battles")
		if canViewAll != nil && canViewAll.(bool) {
			if warriorIDFilter, err := strconv.ParseUint(warriorIDStr, 10, 32); err == nil {
				warriorID = uint(warriorIDFilter)
			}
		} else {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Only emperors and kings can view other warriors' statistics",
			})
			return
		}
	}

	query := dto.GetBattleStatsQuery{
		WarriorID:  warriorID,
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

// GetBattleLogs godoc
// @Summary Get battle logs from Redis
// @Description Get real-time battle logs stored in Redis. Returns all battle events including attacks, critical hits, and battle state changes.
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Battle ID"
// @Param limit query int false "Limit number of logs (default 100, max 1000)"
// @Param from_turn query int false "Start turn number"
// @Param to_turn query int false "End turn number"
// @Success 200 {object} map[string]interface{} "logs: []BattleLogEntry, count: int"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /battles/{id}/logs [get]
func (h *Handler) GetBattleLogs(c *gin.Context) {
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

	// Check if battle exists and user has access
	query := dto.GetBattleQuery{
		BattleID: objectID,
	}

	battle, lightParts, darkParts, err := h.Service.GetBattle(query)
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

	// RBAC check: Check if user is a participant or has admin access
	user, _ := GetCurrentUser(c)
	userIDStr := fmt.Sprintf("%d", user.UserID)
	hasAccess := false
	
	// Check if user is a participant
	for _, p := range lightParts {
		if p.ParticipantID == userIDStr {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		for _, p := range darkParts {
			if p.ParticipantID == userIDStr {
				hasAccess = true
				break
			}
		}
	}

	// Admin/emperor access
	if !hasAccess && !CheckBattleAccess(c, 0) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view this battle's logs",
		})
		return
	}

	ctx := c.Request.Context()
	var logs []BattleLogEntry

	// Check if turn range query is provided
	fromTurnStr := c.Query("from_turn")
	toTurnStr := c.Query("to_turn")
	
	if fromTurnStr != "" && toTurnStr != "" {
		fromTurn, _ := strconv.Atoi(fromTurnStr)
		toTurn, _ := strconv.Atoi(toTurnStr)
		
		if fromTurn >= 0 && toTurn >= fromTurn {
			logs, err = GetBattleLogsByTurnRange(ctx, objectID, fromTurn, toTurn)
			if err != nil {
				c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error:   "internal_error",
					Message: fmt.Sprintf("Failed to retrieve battle logs: %v", err),
				})
				return
			}
		}
	}

	// If no turn range or if range query failed, get all logs (simplified)
	if logs == nil {
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)
		simpleLogs, err := GetSimpleBattleLogs(ctx, objectID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "internal_error",
				Message: fmt.Sprintf("Failed to retrieve battle logs: %v", err),
			})
			return
		}
		
		// Convert to generic format for response
		logsArray := make([]interface{}, len(simpleLogs))
		for i, logEntry := range simpleLogs {
			logsArray[i] = logEntry
		}

		c.JSON(http.StatusOK, gin.H{
			"battle_id": battleID,
			"logs":      logsArray,
			"count":     len(logsArray),
		})
		return
	}

		c.JSON(http.StatusOK, gin.H{
			"battle_id": battleID,
			"logs":      logs,
			"count":     len(logs),
		})
	}
}

// ReviveDragon godoc
// @Summary Revive a dragon in battle
// @Description Revives a defeated dragon participant if it can still revive (max 3 times)
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ReviveDragonRequest true "Revive dragon data"
// @Success 200 {object} map[string]interface{} "participant: BattleParticipant"
// @Failure 400 {object} dto.ErrorResponse
// @Router /battles/revive-dragon [post]
func (h *Handler) ReviveDragon(c *gin.Context) {
	var req dto.ReviveDragonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	battleID, err := primitive.ObjectIDFromHex(req.BattleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid battle ID format",
		})
		return
	}

	participant, err := h.Service.ReviveDragonInBattle(c.Request.Context(), battleID, req.DragonParticipantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "revival_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"participant": dto.ToParticipantResponse(participant),
		"message":    fmt.Sprintf("Dragon %s revived successfully", participant.Name),
	})
}

// DarkEmperorJoinBattle godoc
// @Summary Dark Emperor joins battle during crisis
// @Description Allows Dark Emperor to join an ongoing battle when dragon needs crisis intervention (before 3rd revival)
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.DarkEmperorJoinBattleRequest true "Join battle data"
// @Success 200 {object} map[string]interface{} "participant: BattleParticipant"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /battles/dark-emperor-join [post]
func (h *Handler) DarkEmperorJoinBattle(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if user.Role != "dark_emperor" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only dark emperor can join battle during crisis",
		})
		return
	}

	var req dto.DarkEmperorJoinBattleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Verify username matches authenticated user
	if req.DarkEmperorUsername != user.Username {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "username does not match authenticated user",
		})
		return
	}

	battleID, err := primitive.ObjectIDFromHex(req.BattleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid battle ID format",
		})
		return
	}

	participant, err := h.Service.DarkEmperorJoinBattle(c.Request.Context(), battleID, req.DarkEmperorUsername, req.DarkEmperorUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "join_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"participant": dto.ToParticipantResponse(participant),
		"message":    fmt.Sprintf("Dark Emperor %s joined the battle!", user.Username),
	})
}

// SacrificeDragon godoc
// @Summary Sacrifice dragon and revive all dead enemies
// @Description Dark Emperor can sacrifice a dragon (before 3rd revival) to revive all defeated enemies in battle
// @Tags battles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SacrificeDragonRequest true "Sacrifice dragon data"
// @Success 200 {object} map[string]interface{} "revived_count: int, message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /battles/sacrifice-dragon [post]
func (h *Handler) SacrificeDragon(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if user.Role != "dark_emperor" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only dark emperor can sacrifice dragon",
		})
		return
	}

	var req dto.SacrificeDragonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Verify username matches authenticated user
	if req.DarkEmperorUsername != user.Username {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "username does not match authenticated user",
		})
		return
	}

	battleID, err := primitive.ObjectIDFromHex(req.BattleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid battle ID format",
		})
		return
	}

	revivedCount, multipliedCount, err := h.Service.SacrificeDragonAndReviveEnemies(c.Request.Context(), battleID, req.DragonParticipantID, req.DarkEmperorUsername)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "sacrifice_failed",
			Message: err.Error(),
		})
		return
	}

	totalAffected := revivedCount + multipliedCount
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"revived_count":     revivedCount,
		"multiplied_count": multipliedCount,
		"total_affected":   totalAffected,
		"message":          fmt.Sprintf("Dragon sacrificed! %d enemies revived, %d new enemies created (total: %d).", revivedCount, multipliedCount, totalAffected),
	})

