package warrior

import (
	"net/http"

	"network-sec-micro/internal/warrior/dto"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for warrior service inst
type Handler struct {
	service *Service
}

// NewHandler creates a new handler instance
func NewHandler() *Handler {
	return &Handler{
		service: NewService(),
	}
}

// Login handles warrior login
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Use existing Login function from auth.go
	loginReq := LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	response, err := Login(loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "authentication_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token: response.Token,
		Warrior: dto.WarriorResponse{
			ID:        response.Warrior.ID,
			Username:  response.Warrior.Username,
			Email:     response.Warrior.Email,
			Role:      string(response.Warrior.Role),
			CreatedAt: response.Warrior.CreatedAt,
			UpdatedAt: response.Warrior.UpdatedAt,
		},
	})
}

// GetProfile returns the current warrior's profile
func (h *Handler) GetProfile(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.WarriorResponse{
		ID:        warrior.ID,
		Username:  warrior.Username,
		Email:     warrior.Email,
		Role:      string(warrior.Role),
		CreatedAt: warrior.CreatedAt,
		UpdatedAt: warrior.UpdatedAt,
	})
}

// GetWarriors returns all warriors (King only)
func (h *Handler) GetWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.IsKing() {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only king can access this resource",
		})
		return
	}

	query := dto.GetAllWarriorsQuery{
		Limit:  100,
		Offset: 0,
	}

	warriors, count, err := h.service.GetAllWarriors(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	warriorResponses := make([]dto.WarriorResponse, len(warriors))
	for i, w := range warriors {
		warriorResponses[i] = dto.WarriorResponse{
			ID:        w.ID,
			Username:  w.Username,
			Email:     w.Email,
			Role:      string(w.Role),
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"warriors": warriorResponses,
		"count":    count,
	})
}

// GetKnightWarriors returns all knights (accessible by Knight and King)
func (h *Handler) GetKnightWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	query := dto.GetWarriorsByRoleQuery{
		Role: string(RoleKnight),
	}

	knights, err := h.service.GetWarriorsByRole(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	warriorResponses := make([]dto.WarriorResponse, len(knights))
	for i, k := range knights {
		warriorResponses[i] = dto.WarriorResponse{
			ID:        k.ID,
			Username:  k.Username,
			Email:     k.Email,
			Role:      string(k.Role),
			CreatedAt: k.CreatedAt,
			UpdatedAt: k.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.WarriorsListResponse{
		Role:     string(warrior.Role),
		Warriors: warriorResponses,
		Count:    len(warriorResponses),
	})
}

// GetArcherWarriors returns all archers (accessible by Archer and King)
func (h *Handler) GetArcherWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	query := dto.GetWarriorsByRoleQuery{
		Role: string(RoleArcher),
	}

	archers, err := h.service.GetWarriorsByRole(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	warriorResponses := make([]dto.WarriorResponse, len(archers))
	for i, a := range archers {
		warriorResponses[i] = dto.WarriorResponse{
			ID:        a.ID,
			Username:  a.Username,
			Email:     a.Email,
			Role:      string(a.Role),
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.WarriorsListResponse{
		Role:     string(warrior.Role),
		Warriors: warriorResponses,
		Count:    len(warriorResponses),
	})
}

// GetMageWarriors returns all mages (accessible by Mage and King)
func (h *Handler) GetMageWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	query := dto.GetWarriorsByRoleQuery{
		Role: string(RoleMage),
	}

	mages, err := h.service.GetWarriorsByRole(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	warriorResponses := make([]dto.WarriorResponse, len(mages))
	for i, m := range mages {
		warriorResponses[i] = dto.WarriorResponse{
			ID:        m.ID,
			Username:  m.Username,
			Email:     m.Email,
			Role:      string(m.Role),
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.WarriorsListResponse{
		Role:     string(warrior.Role),
		Warriors: warriorResponses,
		Count:    len(warriorResponses),
	})
}