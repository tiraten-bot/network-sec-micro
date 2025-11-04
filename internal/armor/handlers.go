package armor

import (
    "context"
    "net/http"

    "network-sec-micro/internal/armor/dto"
    "network-sec-micro/pkg/validator"

    "github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for armor service
type Handler struct { Service *Service }

// NewHandler creates a new handler instance
func NewHandler(service *Service) *Handler { return &Handler{Service: service} }

// CreateArmor godoc
// @Summary Create armor
// @Description Create a new armor (Light Emperor/King only). Legendary armors cannot be created.
// @Tags armors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateArmorRequest true "Armor creation data"
// @Success 201 {object} dto.ArmorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /armors [post]
func (h *Handler) CreateArmor(c *gin.Context) {
    user, err := GetCurrentUser(c)
    if err != nil { c.JSON(401, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()}); return }
    if user.Role != "light_emperor" && user.Role != "light_king" {
        c.JSON(403, dto.ErrorResponse{Error: "forbidden", Message: "only light emperor or light king can create armors"}); return
    }
    var req dto.CreateArmorRequest
    if !validator.ValidateRequest(c, &req) { return }
    if req.Type == "legendary" { c.JSON(400, dto.ErrorResponse{Error: "invalid_type", Message: "legendary armors cannot be created"}); return }
    cmd := dto.CreateArmorCommand{ Name: req.Name, Description: req.Description, Type: req.Type, Defense: req.Defense, HPBonus: req.HPBonus, Price: req.Price, MaxDurability: req.MaxDurability, CreatedBy: user.Username }
    a, err := h.Service.CreateArmor(context.Background(), cmd)
    if err != nil { c.JSON(400, dto.ErrorResponse{Error: "creation_failed", Message: err.Error()}); return }
    c.JSON(201, dto.ArmorResponse{ ID: a.ID, Name: a.Name, Description: a.Description, Type: string(a.Type), Defense: a.Defense, HPBonus: a.HPBonus, Price: a.Price, CreatedBy: a.CreatedBy, OwnedBy: a.OwnedBy, Durability: a.Durability, MaxDurability: a.MaxDurability, IsBroken: a.IsBroken, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt })
}

// GetArmors godoc
// @Summary List all armors
// @Tags armors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Armor type filter (common, rare, legendary)"
// @Success 200 {object} dto.ArmorsListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /armors [get]
func (h *Handler) GetArmors(c *gin.Context) {
    query := dto.GetArmorsByTypeRequest{}
    _ = c.ShouldBindQuery(&query)
    q := dto.GetArmorsQuery{ Type: query.Type }
    list, err := h.Service.GetArmors(context.Background(), q)
    if err != nil { c.JSON(500, dto.ErrorResponse{Error: "internal_error", Message: err.Error()}); return }
    resp := make([]dto.ArmorResponse, len(list))
    for i, a := range list { resp[i] = dto.ArmorResponse{ ID: a.ID, Name: a.Name, Description: a.Description, Type: string(a.Type), Defense: a.Defense, HPBonus: a.HPBonus, Price: a.Price, CreatedBy: a.CreatedBy, OwnedBy: a.OwnedBy, Durability: a.Durability, MaxDurability: a.MaxDurability, IsBroken: a.IsBroken, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt } }
    c.JSON(http.StatusOK, dto.ArmorsListResponse{ Armors: resp, Count: len(resp) })
}

// BuyArmor godoc
// @Summary Buy armor
// @Description Purchase an armor. Triggers Kafka event for coin deduction via gRPC.
// @Tags armors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BuyArmorRequest true "Armor purchase data"
// @Success 200 {object} map[string]string "message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /armors/buy [post]
func (h *Handler) BuyArmor(c *gin.Context) {
    user, err := GetCurrentUser(c)
    if err != nil { c.JSON(401, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()}); return }
    var req dto.BuyArmorRequest
    if !validator.ValidateRequest(c, &req) { return }
    cmd := dto.BuyArmorCommand{ ArmorID: req.ArmorID, BuyerRole: user.Role, BuyerID: user.Username, BuyerUsername: user.Username, BuyerUserID: user.UserID, OwnerType: req.OwnerType }
    if err := h.Service.BuyArmor(context.Background(), cmd); err != nil { c.JSON(400, dto.ErrorResponse{Error: "purchase_failed", Message: err.Error()}); return }
    c.JSON(http.StatusOK, gin.H{"message": "armor purchased successfully"})
}

// GetMyArmors godoc
// @Summary Get my armors
// @Tags armors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ArmorsListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /armors/my-armors [get]
func (h *Handler) GetMyArmors(c *gin.Context) {
    user, err := GetCurrentUser(c)
    if err != nil { c.JSON(401, dto.ErrorResponse{Error: "unauthorized", Message: err.Error()}); return }
    list, err := h.Service.GetArmors(context.Background(), dto.GetArmorsQuery{ OwnedBy: user.Username })
    if err != nil { c.JSON(500, dto.ErrorResponse{Error: "internal_error", Message: err.Error()}); return }
    resp := make([]dto.ArmorResponse, len(list))
    for i, a := range list { resp[i] = dto.ArmorResponse{ ID: a.ID, Name: a.Name, Description: a.Description, Type: string(a.Type), Defense: a.Defense, HPBonus: a.HPBonus, Price: a.Price, CreatedBy: a.CreatedBy, OwnedBy: a.OwnedBy, Durability: a.Durability, MaxDurability: a.MaxDurability, IsBroken: a.IsBroken, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt } }
    c.JSON(http.StatusOK, dto.ArmorsListResponse{ Armors: resp, Count: len(resp) })
}


