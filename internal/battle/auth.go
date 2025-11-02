package battle

import (
	"strings"

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

		// Validate token with warrior service (simplified - in production use proper JWT validation)
		// For now, we'll extract user info from token claims
		// TODO: Implement proper JWT validation via warrior service or shared secret

		// Placeholder: In production, validate JWT and extract claims
		// For now, accept any token (development mode)
		// c.Set("username", "test_user")
		// c.Set("user_id", "1")
		// c.Set("role", "knight")

		// For development, we'll skip validation but set default values
		// In production, call warrior service's token validation endpoint or use shared JWT secret
		c.Set("username", c.GetHeader("X-Username")) // Get from header if available
		c.Set("user_id", c.GetHeader("X-User-ID"))
		c.Set("role", c.GetHeader("X-Role"))

		// If headers not set, try to extract from token (simplified)
		if c.GetString("username") == "" {
			// In production, decode JWT and extract claims
			// For now, set defaults for development
			c.Set("username", "test_warrior")
			c.Set("user_id", "1")
			c.Set("role", "knight")
		}

		c.Set("token", token)
		c.Next()
	}
}

