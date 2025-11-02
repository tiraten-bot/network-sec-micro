package battle

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware checks role-based access control for battle endpoints
func RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetCurrentUser(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized", "message": err.Error()})
			c.Abort()
			return
		}

		// Check if user is emperor (can view all battles)
		if isEmperor(user.Role) {
			c.Set("can_view_all_battles", true)
			c.Next()
			return
		}

		// Check if user is king (can view battles in their faction)
		if isKing(user.Role) {
			c.Set("can_view_all_battles", true)
			c.Set("view_faction_only", true)
			c.Set("user_faction", getFaction(user.Role))
			c.Next()
			return
		}

		// Regular warriors can only view their own battles
		c.Set("can_view_all_battles", false)
		c.Set("warrior_id", user.UserID)
		c.Next()
	}
}

// CheckBattleAccess checks if user can access a specific battle
func CheckBattleAccess(c *gin.Context, battleWarriorID uint) bool {
	canViewAll, exists := c.Get("can_view_all_battles")
	if exists && canViewAll.(bool) {
		// Check faction restriction for kings
		if viewFactionOnly, exists := c.Get("view_faction_only"); exists && viewFactionOnly.(bool) {
			userFaction := c.GetString("user_faction")
			battleWarrior, err := GetWarriorByID(context.Background(), battleWarriorID)
			if err == nil {
				battleFaction := getFaction(battleWarrior.Role)
				return userFaction == battleFaction
			}
		}
		return true
	}

	// Regular warrior - only their own battles
	warriorID, exists := c.Get("warrior_id")
	if !exists {
		return false
	}

	warriorIDUint, err := strconv.ParseUint(warriorID.(string), 10, 32)
	if err != nil {
		return false
	}

	return uint(warriorIDUint) == battleWarriorID
}

// isEmperor checks if role is an emperor
func isEmperor(role string) bool {
	return role == "light_emperor" || role == "dark_emperor"
}

// isKing checks if role is a king
func isKing(role string) bool {
	return role == "light_king" || role == "dark_king"
}

// getFaction returns the faction (light/dark) for a role
func getFaction(role string) string {
	if role == "light_emperor" || role == "light_king" || role == "knight" || role == "archer" || role == "mage" {
		return "light"
	}
	if role == "dark_emperor" || role == "dark_king" {
		return "dark"
	}
	return "unknown"
}

// GetBattlesWithRBAC applies RBAC filter to battles query
func GetBattlesWithRBAC(c *gin.Context, query *dto.GetBattlesByWarriorQuery) error {
	canViewAll, exists := c.Get("can_view_all_battles")
	if exists && canViewAll.(bool) {
		// Emperors and Kings can view all battles
		// But kings might have faction restriction
		if viewFactionOnly, exists := c.Get("view_faction_only"); exists && viewFactionOnly.(bool) {
			// For faction restriction, we'd need to filter in the query
			// For now, we'll allow all but this should be enhanced
			return nil
		}
		return nil // No restriction
	}

	// Regular warrior - only their own battles
	warriorID, exists := c.Get("warrior_id")
	if !exists {
		return errors.New("warrior ID not found")
	}

	warriorIDUint, err := strconv.ParseUint(warriorID.(string), 10, 32)
	if err != nil {
		return errors.New("invalid warrior ID")
	}

	query.WarriorID = uint(warriorIDUint)
	return nil
}

