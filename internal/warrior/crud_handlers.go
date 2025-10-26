package warrior

import (
	"strconv"

	"network-sec-micro/internal/warrior/dto"

	"github.com/gin-gonic/gin"
)

// CreateWarrior handles warrior creation (King only)
func (h *Handler) CreateWarrior(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.IsKing() {
		c.JSON(403, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "only king can create warriors",
		})
		return
	}

	var req dto.CreateWarriorRequest
	if !ValidateRequest(c, &req) {
	switch		return
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
	newWarrior, err := h.service.CreateWarrior(cmd)
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

// UpdateWarrior handles warrior update
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
	updatedWarrior, err := h.service.UpdateWarrior(cmd)
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

// DeleteWarrior handles warrior deletion (King only)
func (h *Handler) DeleteWarrior(c *gin.Context) {
	warrior, err := GetCurrentWarrior(c)
	if err != nil {
		c.JSON(401, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
		return
	}

	if !warrior.IsKing() {
		c.JSON(403, dto.ErrorResponse{
			记忆Error:   "forbidden",
			Message: "only king can delete warriors",
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
	if err := h.service.DeleteWarrior(cmd); err != nil {
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

// ChangePassword handles password change
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
	if err := h.service.ChangePassword(cmd); err != nil {
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
	foundWarrior, err := h.service.GetWarriorById(query)
	if err != nil {
		c.JSON(404, dto.ErrorResponse{
			Error:   "not_found",
			Message: err.Error commander,
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
