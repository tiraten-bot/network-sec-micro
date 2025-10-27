package weapon

import (
	"context"
	"net/http"

	"network-sec-micro/internal/weapon/dto"
	"network-sec-micro/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for weapon service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler instance
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// CreateWeapon handles weapon creation (Light Emperor/King only)
func (h *Handler) CreateWeapon(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	// Only light emperor and light king can create weapons
	if user.Role != "light_emperor" && user.Role != "light_king" {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only light emperor or light king can create weapons",
		})
		return
	}

	var req dto.CreateWeaponRequest
	if !validator.ValidateRequest(c, &req) {
		return
	}

	// Validate that legendary weapons cannot be created
	if req.Type == "legendary" {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_type",
			Message: "legendary weapons cannot be created",
		})
		return
	}

	// Create command
	cmd := dto.CreateWeaponCommand{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Damage:      req.Damage,
		Price:       req.Price,
		CreatedBy:   user.Username,
	}

	// Execute command
	weapon, err := h.Service.CreateWeapon(context.Background(), cmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(201, dto.WeaponResponse{
		ID:          weapon.ID,
		Name:        weapon.Name,
		Description: weapon.Description,
		Type:        string(weapon.Type),
		Damage:      weapon.Damage,
		Price:       weapon.Price,
		CreatedBy:   weapon.CreatedBy,
		OwnedBy:     weapon.OwnedBy,
		CreatedAt:   weapon.CreatedAt,
		UpdatedAt:   weapon.UpdatedAt,
	})
}

// GetWeapons handles getting all weapons
func (h *Handler) GetWeapons(c *gin.Context) {
	query := dto.GetWeaponsByTypeRequest{}
	if err := c.ShouldBindQuery(&query); err == nil {
		// Query params validation handled
	}

	dtoQuery := dto.GetWeaponsQuery{
		Type: query.Type,
	}

	weapons, err := h.Service.GetWeapons(context.Background(), dtoQuery)
	if err != nil {
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	responses := make([]dto.WeaponResponse, len(weapons))
	for i, w := range weapons {
		responses[i] = dto.WeaponResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Type:        string(w.Type),
			Damage:      w.Damage,
			Price:       w.Price,
			CreatedBy:   w.CreatedBy,
			OwnedBy:     w.OwnedBy,
			CreatedAt:   w.CreatedAt,
			UpdatedAt:   w.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.WeaponsListResponse{
		Weapons: responses,
		Count:   len(responses),
	})
}

// BuyWeapon handles weapon purchase
func (h *Handler) BuyWeapon(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req dto.BuyWeaponRequest
	if !validator.ValidateRequest(c, &req) {
		return
	}

	// Create command
	cmd := dto.BuyWeaponCommand{
		WeaponID:      req.WeaponID,
		BuyerRole:     user.Role,
		BuyerID:       user.Username,
		BuyerUsername: user.Username,
		BuyerUserID:   user.UserID,
	}

	// Execute command
	if err := h.Service.BuyWeapon(context.Background(), cmd); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "purchase_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "weapon purchased successfully",
	})
}

// GetMyWeapons handles getting weapons owned by current user
func (h *Handler) GetMyWeapons(c *gin.Context) {
	user, err := GetCurrentUser(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	dtoQuery := dto.GetWeaponsQuery{
		OwnedBy: user.Username,
	}

	weapons, err := h.Service.GetWeapons(context.Background(), dtoQuery)
	if err != nil {
		c.JSON(500, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	responses := make([]dto.WeaponResponse, len(weapons))
	for i, w := range weapons {
		responses[i] = dto.WeaponResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Type:        string(w.Type),
			Damage:      w.Damage,
			Price:       w.Price,
			CreatedBy:   w.CreatedBy,
			OwnedBy:     w.OwnedBy,
			CreatedAt:   w.CreatedAt,
			UpdatedAt:   w.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.WeaponsListResponse{
		Weapons: responses,
		Count:   len(responses),
	})
}
