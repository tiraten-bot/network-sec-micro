package warrior

import (
	"network-sec-micro/internal/warrior/dto"

	"github.com/gin-gonic/gin"
)

// CreateKing handles king creation (Emperor only)
func (h *Handler) CreateKing(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.CanCreateKings() {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only emperors can create kings",
		})
		return
	}

	var req dto.CreateKingRequest
	if !ValidateRequest(c, &req) {
		return
	}

	// Validate king role
	validRole := Role(req.Role)
	if warrior.Role == RoleLightEmperor && validRole != RoleLightKing {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_role",
			Message: "light emperor can only create light king",
		})
		return
	}
	if warrior.Role == RoleDarkEmperor && validRole != RoleDarkKing {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_role",
			Message: "dark emperor can only create dark king",
		})
		return
	}

	// Create command
	cmd := dto.CreateKingCommand{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Role:      req.Role,
		CreatedBy: warrior.ID,
	}

	// Execute command (convert CreateKingCommand to CreateWarriorCommand)
	warriorCmd := dto.CreateWarriorCommand{
		Username:  cmd.Username,
		Email:     cmd.Email,
		Password:  cmd.Password,
		Role:      cmd.Role,
		CreatedBy: cmd.CreatedBy,
	}
	newKing, err := h.Service.CreateWarrior(warriorCmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(201, dto.WarriorResponse{
		ID:        newKing.ID,
		Username:  newKing.Username,
		Email:     newKing.Email,
		Role:      string(newKing.Role),
		CreatedAt: newKing.CreatedAt,
		UpdatedAt: newKing.UpdatedAt,
	})
}

// GetKings handles getting all kings (Emperors only)
func (h *Handler) GetKings(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.IsEmperor() {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only emperors can view kings",
		})
		return
	}

	query := dto.GetAllWarriorsQuery{
		Limit:  100,
		Offset: 0,
	}

	warriors, _, err := h.Service.GetAllWarriors(query)
	if err != nil {
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	// Filter to show only kings
	var kings []Warrior
	for _, w := range warriors {
		if w.IsKing() {
			kings = append(kings, w)
		}
	}

	kingResponses := make([]dto.WarriorResponse, len(kings))
	for i, k := range kings {
		kingResponses[i] = dto.WarriorResponse{
			ID:        k.ID,
			Username:  k.Username,
			Email:     k.Email,
			Role:      string(k.Role),
			CreatedAt: k.CreatedAt,
			UpdatedAt: k.UpdatedAt,
		}
	}

	c.JSON(200, gin.H{
		"kings": kingResponses,
		"count": len(kings),
	})
}
