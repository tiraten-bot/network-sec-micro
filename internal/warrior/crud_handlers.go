package warrior

import (
	"strconv"

	"network-sec-micro/internal/warrior/dto"

	"github.com/gin-gonic/gin"
)

// CreateWarrior godoc
// @Summary Create warrior
// @Description Create a new warrior (Light Emperor/King only)
// @Tags warriors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateWarriorRequest true "Warrior creation data"
// @Success 201 {object} dto.WarriorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /warriors [post]
func (h *Handler) CreateWarrior(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.CanCreateWarriors() {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only light emperor or light king can create warriors",
		})
		return
	}

	var req dto.CreateWarriorRequest
	if !ValidateRequest(c, &req) {
		return
	}

	// Validate role
	validRole := Role(req.Role)
	if validRole != RoleKnight && validRole != RoleArcher && validRole != RoleMage {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_role",
			Message: "role must be one of: knight, archer, mage",
		})
		return
	}

	// Create command
	cmd := dto.CreateWarriorCommand{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Role:      req.Role,
		CreatedBy: warrior.ID,
	}

	// Execute command
	newWarrior, err := h.Service.CreateWarrior(cmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(201, dto.WarriorResponse{
		ID:        newWarrior.ID,
		Username:  newWarrior.Username,
		Email:     newWarrior.Email,
		Role:      string(newWarrior.Role),
		CreatedAt: newWarrior.CreatedAt,
		UpdatedAt: newWarrior.UpdatedAt,
	})
}

// UpdateWarrior godoc
// @Summary Update warrior
// @Description Update warrior information (own profile or King can update any)
// @Tags warriors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Warrior ID"
// @Param request body dto.UpdateWarriorRequest true "Update data"
// @Success 200 {object} dto.WarriorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /warriors/{id} [put]
func (h *Handler) UpdateWarrior(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	// Get warrior ID from path
	warriorIDParam := c.Param("id")
	warriorID, err := strconv.ParseUint(warriorIDParam, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "invalid warrior ID",
		})
		return
	}

	// Check permissions
	if !warrior.IsKing() && warrior.ID != uint(warriorID) {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "you can only update your own profile",
		})
		return
	}

	var req dto.UpdateWarriorRequest
	if !ValidateRequest(c, &req) {
		return
	}

	// Build command
	cmd := dto.UpdateWarriorCommand{
		WarriorID: uint(warriorID),
		UpdatedBy: warrior.ID,
	}

	if req.Email != "" {
		cmd.Email = &req.Email
	}
	if req.Role != "" {
		cmd.Role = &req.Role
	}

	// Execute command
	updatedWarrior, err := h.Service.UpdateWarrior(cmd)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.WarriorResponse{
		ID:        updatedWarrior.ID,
		Username:  updatedWarrior.Username,
		Email:     updatedWarrior.Email,
		Role:      string(updatedWarrior.Role),
		CreatedAt: updatedWarrior.CreatedAt,
		UpdatedAt: updatedWarrior.UpdatedAt,
	})
}

// DeleteWarrior godoc
// @Summary Delete warrior
// @Description Delete a warrior (Light Emperor/King only, cannot delete self)
// @Tags warriors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Warrior ID"
// @Success 200 {object} map[string]string "message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /warriors/{id} [delete]
func (h *Handler) DeleteWarrior(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.CanCreateWarriors() {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only light emperor or light king can delete warriors",
		})
		return
	}

	// Get warrior ID from path
	warriorIDParam := c.Param("id")
	warriorID, err := strconv.ParseUint(warriorIDParam, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "invalid warrior ID",
		})
		return
	}

	// Cannot delete yourself
	if warrior.ID == uint(warriorID) {
		c.JSON(400, dto.ErrorResponse{
			Error:   "cannot_delete_self",
			Message: "cannot delete yourself",
		})
		return
	}

	// Create command
	cmd := dto.DeleteWarriorCommand{
		WarriorID: uint(warriorID),
		DeletedBy: warrior.ID,
	}

	// Execute command
	if err := h.Service.DeleteWarrior(cmd); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "warrior deleted successfully",
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change authenticated warrior's password
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]string "message: string"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /profile/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	var req dto.ChangePasswordRequest
	if !ValidateRequest(c, &req) {
		return
	}

	// Create command
	cmd := dto.ChangePasswordCommand{
		WarriorID:   warrior.ID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		ChangedBy:   warrior.ID,
	}

	// Execute command
	if err := h.Service.ChangePassword(cmd); err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "password_change_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "password changed successfully",
	})
}

// GetWarriorById handles getting a single warrior by ID
func (h *Handler) GetWarriorById(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	// Get warrior ID from path
	warriorIDParam := c.Param("id")
	warriorID, err := strconv.ParseUint(warriorIDParam, 10, 32)
	if err != nil {
		c.JSON(400, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "invalid warrior ID",
		})
		return
	}

	// Check permissions - users can only see themselves or must be king
	if !warrior.IsKing() && warrior.ID != uint(warriorID) {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "you can only view your own profile",
		})
		return
	}

	// Create query
	query := dto.GetWarriorQuery{
		WarriorID: uint(warriorID),
	}

	// Execute query
	foundWarrior, err := h.Service.GetWarriorById(query)
	if err != nil {
		c.JSON(404, dto.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, dto.WarriorResponse{
		ID:        foundWarrior.ID,
		Username:  foundWarrior.Username,
		Email:     foundWarrior.Email,
		Role:      string(foundWarrior.Role),
		CreatedAt: foundWarrior.CreatedAt,
		UpdatedAt: foundWarrior.UpdatedAt,
	})
}
