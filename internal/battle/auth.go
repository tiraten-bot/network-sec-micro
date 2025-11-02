package battle

import (
	"errors"
	"strconv"
	"strings"

	"network-sec-micro/pkg/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens from warrior service
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "unauthorized", "message": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "unauthorized", "message": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized", "message": "invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("username", claims.Username)
		c.Set("user_id", strconv.FormatUint(uint64(claims.UserID), 10))
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GetCurrentUser returns the current user from context
func GetCurrentUser(c *gin.Context) (*User, error) {
	username := c.GetString("username")
	if username == "" {
		return nil, errors.New("username not found in context")
	}

	userIDStr := c.GetString("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	return &User{
		Username: username,
		UserID:   uint(userID),
		Role:     c.GetString("role"),
	}, nil
}


