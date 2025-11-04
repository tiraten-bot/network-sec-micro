package battle

import (
	"context"
	"errors"
	"strconv"

	"network-sec-micro/internal/battle/dto"

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
	user, err := GetCurrentUser(c)
	if err != nil {
		return false
	}

	// Emperors can view all battles
	if isEmperor(user.Role) {
		return true
	}

	// Kings can view battles in their faction
	if isKing(user.Role) {
		userFaction := getFaction(user.Role)
		// Get battle warrior to check faction
		battleWarrior, err := GetWarriorByID(context.Background(), battleWarriorID)
		if err == nil {
			battleFaction := getFaction(battleWarrior.Role)
			// Kings cannot see emperor battles (cross-faction restriction)
			if isEmperor(battleWarrior.Role) {
				return false // Kings cannot see emperor battles
			}
			return userFaction == battleFaction
		}
		// If we can't get warrior info, deny access
		return false
	}

	// Regular warriors can only see their own battles
	return user.UserID == battleWarriorID
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
		// Emperors can view all battles (no filter)
		// Kings can view battles in their faction only
		viewFactionOnly, factionExists := c.Get("view_faction_only")
		if factionExists && viewFactionOnly.(bool) {
			// Kings see only their faction battles
			// For now, we'll let them see all but in production,
			// we'd filter by warrior faction in the service layer
			// Setting warriorID to 0 means "all" in our query
			query.WarriorID = 0
			return nil
		}
		// Emperors - no restriction, can see all
		query.WarriorID = 0 // 0 means "all battles"
		return nil
	}

	// Regular warrior - only their own battles
	warriorID, exists := c.Get("warrior_id")
	if !exists {
		return errors.New("warrior ID not found")
	}

	warriorIDStr, ok := warriorID.(string)
	if !ok {
		return errors.New("invalid warrior ID type")
	}

	warriorIDUint, err := strconv.ParseUint(warriorIDStr, 10, 32)
	if err != nil {
		return errors.New("invalid warrior ID format")
	}

	query.WarriorID = uint(warriorIDUint)
	return nil
}

