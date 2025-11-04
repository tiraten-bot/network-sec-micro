package armor

import (
	"errors"
	"strings"

	"network-sec-micro/pkg/auth"

	"github.com/gin-gonic/gin"
)

// User represents authenticated user info from JWT
type User struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AuthMiddleware validates JWT token and sets user in context
func AuthMiddleware() gin.HandlerFunc {
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

		token := parts[1]
		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user in context
		user := User{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}
		c.Set("user", &user)
		c.Next()
	}
}

// GetCurrentUser returns the current user from context
func GetCurrentUser(c *gin.Context) (*User, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}

	user, ok := userInterface.(*User)
	if !ok {
		return nil, errors.New("invalid user data")
	}

	return user, nil
}

