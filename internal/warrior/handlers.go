package warrior

import (
	"net/http"

	"network-sec-micro/internal/warrior/dto"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for warrior service
type Handler struct {
	Service *Service
}

// NewHandler creates a new handler instance
func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

// Login godoc
// @Summary Login warrior
// @Description Authenticate warrior and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /login [post]
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
			Title:     response.Warrior.Title,
			CreatedAt: response.Warrior.CreatedAt,
			UpdatedAt: response.Warrior.UpdatedAt,
		},
	})
}

// GetProfile godoc
// @Summary Get current warrior profile
// @Description Get the authenticated warrior's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.WarriorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile [get]
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
		Title:     warrior.Title,
		CreatedAt: warrior.CreatedAt,
		UpdatedAt: warrior.UpdatedAt,
	})
}

// GetWarriors godoc
// @Summary List all warriors
// @Description Get list of all warriors (Light Emperor/King only)
// @Tags warriors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "warriors: []WarriorResponse, count: int64"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /warriors [get]
func (h *Handler) GetWarriors(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.CanCreateWarriors() {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only light emperor or light king can access this resource",
		})
		return
	}

	query := dto.GetAllWarriorsQuery{
		Limit:  100,
		Offset: 0,
	}

	warriors, count, err := h.Service.GetAllWarriors(query)
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
			Title:     w.Title,
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

	knights, err := h.Service.GetWarriorsByRole(query)
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
			Title:     k.Title,
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

	archers, err := h.Service.GetWarriorsByRole(query)
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
			Title:     a.Title,
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

	mages, err := h.Service.GetWarriorsByRole(query)
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
			Title:     m.Title,
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

// GetMyKilledMonsters godoc
// @Summary Get warrior's killed monsters
// @Description List all monsters killed by the authenticated warrior
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "kills: []KilledMonster, count: int64"
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile/kills [get]
func (h *Handler) GetMyKilledMonsters(c *gin.Context) {
    warrior, err := GetCurrentWarrior(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()})
        return
    }
    kills, count, err := h.Service.GetKilledMonsters(warrior.ID, 100, 0)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "kills": kills,
        "count": count,
    })
}

// GetMyStrongestKill godoc
// @Summary Get strongest killed monster
// @Description Get the strongest monster killed by authenticated warrior
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "strongest: KilledMonster"
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile/strongest-kill [get]
func (h *Handler) GetMyStrongestKill(c *gin.Context) {
    warrior, err := GetCurrentWarrior(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()})
        return
    }
    km, err := h.Service.GetStrongestKilledMonster(warrior.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
        return
    }
    if km == nil {
        c.JSON(http.StatusOK, gin.H{"strongest": nil})
        return
    }
    c.JSON(http.StatusOK, gin.H{"strongest": km})
}