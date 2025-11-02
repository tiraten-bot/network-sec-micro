package dragon

import (
	"errors"

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

// CreateDragon godoc
// @Summary Create dragon
// @Description Create a new dragon. Creator info is extracted from JWT token.
// @Tags dragons
// @Accept json
// @Produce json
// @Param request body dto.CreateDragonRequest true "Dragon creation data"
// @Success 201 {object} dto.CreateDragonResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /dragons [post]
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

	// Convert Dragon to dto.Dragon
	dtoDragon := &dto.Dragon{
		ID:                        dragon.ID,
		Name:                      dragon.Name,
		Type:                      string(dragon.Type),
		Level:                     dragon.Level,
		Health:                    dragon.Health,
		MaxHealth:                 dragon.MaxHealth,
		AttackPower:               dragon.AttackPower,
		Defense:                   dragon.Defense,
		CreatedBy:                 dragon.CreatedBy,
		IsAlive:                   dragon.IsAlive,
		KilledBy:                  dragon.KilledBy,
		RevivalCount:              dragon.RevivalCount,
		AwaitingCrisisIntervention: dragon.AwaitingCrisisIntervention,
		CreatedAt:                 dragon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                 dragon.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if dragon.KilledAt != nil {
		killedAtStr := dragon.KilledAt.Format("2006-01-02T15:04:05Z07:00")
		dtoDragon.KilledAt = &killedAtStr
	}

	c.JSON(201, dto.CreateDragonResponse{
		Success: true,
		Dragon:  dtoDragon,
		Message: "Dragon created successfully",
	})
}

// AttackDragon godoc
// @Summary Attack dragon
// @Description Attack a dragon. When health reaches 0, death event is published to Kafka.
// @Tags dragons
// @Accept json
// @Produce json
// @Param id path string true "Dragon ID"
// @Param request body dto.AttackDragonRequest true "Attack data"
// @Success 200 {object} dto.AttackDragonResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /dragons/{id}/attack [post]
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

	// Convert Dragon to dto.Dragon
	dtoDragon := &dto.Dragon{
		ID:          dragon.ID,
		Name:        dragon.Name,
		Type:        string(dragon.Type),
		Level:       dragon.Level,
		Health:      dragon.Health,
		MaxHealth:   dragon.MaxHealth,
		AttackPower: dragon.AttackPower,
		Defense:     dragon.Defense,
		CreatedBy:   dragon.CreatedBy,
		IsAlive:     dragon.IsAlive,
		KilledBy:    dragon.KilledBy,
		CreatedAt:   dragon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   dragon.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if dragon.KilledAt != nil {
		killedAtStr := dragon.KilledAt.Format("2006-01-02T15:04:05Z07:00")
		dtoDragon.KilledAt = &killedAtStr
	}

	c.JSON(200, dto.AttackDragonResponse{
		Success: true,
		Dragon:  dtoDragon,
		Message: "Dragon attacked successfully",
	})
}

// GetDragon godoc
// @Summary Get dragon by ID
// @Description Get dragon details by ID
// @Tags dragons
// @Accept json
// @Produce json
// @Param id path string true "Dragon ID"
// @Success 200 {object} dto.GetDragonResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /dragons/{id} [get]
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

	// Convert Dragon to dto.Dragon
	dtoDragon := &dto.Dragon{
		ID:          dragon.ID,
		Name:        dragon.Name,
		Type:        string(dragon.Type),
		Level:       dragon.Level,
		Health:      dragon.Health,
		MaxHealth:   dragon.MaxHealth,
		AttackPower: dragon.AttackPower,
		Defense:     dragon.Defense,
		CreatedBy:   dragon.CreatedBy,
		IsAlive:     dragon.IsAlive,
		KilledBy:    dragon.KilledBy,
		CreatedAt:   dragon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   dragon.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if dragon.KilledAt != nil {
		killedAtStr := dragon.KilledAt.Format("2006-01-02T15:04:05Z07:00")
		dtoDragon.KilledAt = &killedAtStr
	}

	c.JSON(200, dto.GetDragonResponse{
		Success: true,
		Dragon:  dtoDragon,
	})
}

// GetDragonsByType godoc
// @Summary Get dragons by type
// @Description Get list of dragons filtered by type, optionally only alive ones
// @Tags dragons
// @Accept json
// @Produce json
// @Param type path string true "Dragon type (fire, ice, earth, air)"
// @Param alive query bool false "Filter only alive dragons"
// @Success 200 {object} dto.GetDragonsByTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /dragons/type/{type} [get]
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

	// Convert Dragons to dto.Dragons
	var dtoDragons []dto.Dragon
	for _, dragon := range dragons {
		dtoDragon := dto.Dragon{
			ID:          dragon.ID,
			Name:        dragon.Name,
			Type:        string(dragon.Type),
			Level:       dragon.Level,
			Health:      dragon.Health,
			MaxHealth:   dragon.MaxHealth,
			AttackPower: dragon.AttackPower,
			Defense:     dragon.Defense,
			CreatedBy:   dragon.CreatedBy,
			IsAlive:     dragon.IsAlive,
			KilledBy:    dragon.KilledBy,
			CreatedAt:   dragon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   dragon.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if dragon.KilledAt != nil {
			killedAtStr := dragon.KilledAt.Format("2006-01-02T15:04:05Z07:00")
			dtoDragon.KilledAt = &killedAtStr
		}
		dtoDragons = append(dtoDragons, dtoDragon)
	}

	c.JSON(200, dto.GetDragonsByTypeResponse{
		Success: true,
		Dragons: dtoDragons,
		Count:   len(dtoDragons),
	})
}

// GetDragonsByCreator godoc
// @Summary Get dragons by creator
// @Description Get list of dragons created by a specific warrior, optionally only alive ones
// @Tags dragons
// @Accept json
// @Produce json
// @Param creator path string true "Creator username"
// @Param alive query bool false "Filter only alive dragons"
// @Success 200 {object} dto.GetDragonsByCreatorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /dragons/creator/{creator} [get]
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

	// Convert Dragons to dto.Dragons
	var dtoDragons []dto.Dragon
	for _, dragon := range dragons {
		dtoDragon := dto.Dragon{
			ID:          dragon.ID,
			Name:        dragon.Name,
			Type:        string(dragon.Type),
			Level:       dragon.Level,
			Health:      dragon.Health,
			MaxHealth:   dragon.MaxHealth,
			AttackPower: dragon.AttackPower,
			Defense:     dragon.Defense,
			CreatedBy:   dragon.CreatedBy,
			IsAlive:     dragon.IsAlive,
			KilledBy:    dragon.KilledBy,
			CreatedAt:   dragon.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   dragon.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if dragon.KilledAt != nil {
			killedAtStr := dragon.KilledAt.Format("2006-01-02T15:04:05Z07:00")
			dtoDragon.KilledAt = &killedAtStr
		}
		dtoDragons = append(dtoDragons, dtoDragon)
	}

	c.JSON(200, dto.GetDragonsByCreatorResponse{
		Success: true,
		Dragons: dtoDragons,
		Count:   len(dtoDragons),
	})
}
