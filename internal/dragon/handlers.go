package dragon

import (
	"context"
	"errors"
	"fmt"
	"log"

	"network-sec-micro/internal/dragon/dto"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Handler handles HTTP requests for dragon service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler instance
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// CreateDragon handles dragon creation
func (h *Handler) CreateDragon(c *gin.Context) {
	var req dto.CreateDragonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get creator info from JWT token (simplified for now)
	creatorUsername := c.GetString("username")
	creatorRole := c.GetString("role")

	cmd := dto.CreateDragonCommand{
		Name:        req.Name,
		Type:        req.Type,
		Level:       req.Level,
		CreatedBy:   creatorUsername,
		CreatedByRole: creatorRole,
	}

	dragon, err := h.Service.CreateDragon(cmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(201, dto.CreateDragonResponse{
		Success: true,
		Dragon:  dragon,
		Message: "Dragon created successfully",
	})
}

// AttackDragon handles dragon attack
func (h *Handler) AttackDragon(c *gin.Context) {
	var req dto.AttackDragonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Get attacker info from JWT token
	attackerUsername := c.GetString("username")

	cmd := dto.AttackDragonCommand{
		DragonID:         req.DragonID,
		AttackerUsername: attackerUsername,
	}

	dragon, err := h.Service.AttackDragon(cmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "attack_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.AttackDragonResponse{
		Success: true,
		Dragon:  dragon,
		Message: "Dragon attacked successfully",
	})
}

// GetDragon handles getting dragon by ID
func (h *Handler) GetDragon(c *gin.Context) {
	dragonID := c.Param("id")
	if dragonID == "" {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "dragon ID is required",
		})
		return
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(dragonID)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "invalid dragon ID format",
		})
		return
	}

	query := dto.GetDragonQuery{
		DragonID: objectID,
	}

	dragon, err := h.Service.GetDragon(query)
	if err != nil {
		if errors.Is(err, errors.New("dragon not found")) {
			c.JSON(404, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Dragon not found",
			})
			return
		}
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.GetDragonResponse{
		Success: true,
		Dragon:  dragon,
	})
}

// GetDragonsByType handles getting dragons by type
func (h *Handler) GetDragonsByType(c *gin.Context) {
	dragonType := c.Param("type")
	if dragonType == "" {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "dragon type is required",
		})
		return
	}

	aliveOnly := c.Query("alive") == "true"

	query := dto.GetDragonsByTypeQuery{
		Type:      dragonType,
		AliveOnly: aliveOnly,
	}

	dragons, err := h.Service.GetDragonsByType(query)
	if err != nil {
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.GetDragonsByTypeResponse{
		Success: true,
		Dragons: dragons,
		Count:   len(dragons),
	})
}

// GetDragonsByCreator handles getting dragons by creator
func (h *Handler) GetDragonsByCreator(c *gin.Context) {
	creatorUsername := c.Param("creator")
	if creatorUsername == "" {
		c.JSON(400, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "creator username is required",
		})
		return
	}

	aliveOnly := c.Query("alive") == "true"

	query := dto.GetDragonsByCreatorQuery{
		CreatorUsername: creatorUsername,
		AliveOnly:       aliveOnly,
	}

	dragons, err := h.Service.GetDragonsByCreator(query)
	if err != nil {
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.GetDragonsByCreatorResponse{
		Success: true,
		Dragons: dragons,
		Count:   len(dragons),
	})
}
