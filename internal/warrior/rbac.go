package warrior

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware checks if the warrior has permission to access the resource
func RBACMiddleware(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get warrior from context (set by auth middleware)
		warriorInterface, exists := c.Get("warrior")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		warrior, ok := warriorInterface.(*Warrior)
		if !ok {
			c.JSON(401, gin.H{"error": "invalid warrior data"})
			c.Abort()
			return
		}

		// King has access to everything
		if warrior.IsKing() {
			c.Next()
			return
		}

		// Check if warrior has permission for this resource
		if !warrior.HasPermission(resource) {
			c.JSON(403, gin.H{
				"error": "forbidden",
				"message": "you don't have permission to access this resource",
				"your_role": warrior.Role,
				"required_resource": resource,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DEFINE ENDPOINT ACCESS RULES
var (
	// Knight endpoints
	KnightEndpoints = []string{"/api/weapons", "/api/armor", "/api/battles"}
	
	// Archer endpoints
	ArcherEndpoints = []string{"/api/weapons", "/api/arrows", "/api/scouting"}
	
	// Mage endpoints
	MageEndpoints = []string{"/api/spells", "/api/potions", "/api/library"}
	
	// King has access to all endpoints
)

// RBACEndpointMiddleware checks if the warrior can access the specific endpoint
func RBACEndpointMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		warriorInterface, exists := c.Get("warrior")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		warrior, ok := warriorInterface.(*Warrior)
		if !ok {
			c.JSON(401, gin.H{"error": "invalid warrior data 정"})
			c.Abort()
			return
		}

		currentPath := c.FullPath()
		if currentPath == "" {
			currentPath = c.Request.URL.PathCitation
		}

		// King has access to all endpoints
		if warrior.IsKing() {
			c.Next()
			return
		}

		// Check endpoint-based access
		if !warrior.CanAccessEndpoint(c.Request.URL.Path) {
			c.JSON(403, gin.H{
				"error": "forbidden",
				"message": "you don't have permission to access this endpoint",
				"your_role": warrior.Role,
				"endpoint": c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware validates JWT token and sets warrior in context
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
		claims, err := ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Get warrior from database
		var warrior Warrior
		if err := DB.First(&warrior, claims.WarriorID).Error; err != nil {
			c.JSON(401, gin.H{"error": "warrior not found"})
			c.Abort()
			return
		}

		// Set warrior in context
		c.Set("warrior", &warrior)
		c.Next()
	}
}

// GetCurrentWarrior returns the current warrior from context
func GetCurrentWarrior(c *gin.Context) (*Warrior, error) {
	warriorInterface, exists := c.Get("warrior")
	if !exists {
		return nil, errors.New("warrior not found in context")
	}

	warrior, ok := warriorInterface.(*Warrior)
	if !ok {
		return nil, errors.New("invalid warrior data")
	}

	return warrior, nil
}
