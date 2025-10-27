package coin

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCurrentUserFromContext extracts user info from gin context (set by warrior service)
func GetCurrentUserFromContext(c *gin.Context) (*AuthUser, error) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("user_id not found in context")
	}

	roleInterface, exists := c.Get("role")
	if !exists {
		return nil, errors.New("role not found in context")
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		return nil, errors.New("invalid user_id type")
	}

	role, ok := roleInterface.(string)
	if !ok {
		return nil, errors.New("invalid role type")
	}

	return &AuthUser{
		ID:   userID,
		Role: role,
	}, nil
}

// AuthUser represents authenticated user
type AuthUser struct {
	ID   uint
	Role string
}

// CanViewAllWarriors checks if user can view all warriors' balances
func (u *AuthUser) CanViewAllWarriors() bool {
	return u.Role == "light_king" || u.Role == "light_emperor"
}

// CanViewBalance checks if user can view a specific warrior's balance
func (u *AuthUser) CanViewBalance(warriorID uint) bool {
	// Can view own balance
	if u.ID == warriorID {
		return true
	}
	// Kings and emperors can view all
	return u.CanViewAllWarriors()
}

// RBACMiddleware validates authorization for coin operations
func RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// TODO: Validate JWT token and extract claims
		// For now, get from query or header
		userID := c.GetHeader("X-User-ID")
		role := c.GetHeader("X-User-Role")
		
		if userID == "" || role == "" {
			c.JSON(401, gin.H{"error": "user information required"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

