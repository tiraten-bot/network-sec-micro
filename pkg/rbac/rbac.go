package rbac

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// User represents a user with role information
type User interface {
	GetRole() string
	IsAdmin() bool
}

// ResourceChecker checks if a user has access to a resource
type ResourceChecker interface {
	CanAccessResource(resource string) bool
}

// AdminRole represents the admin role that can access everything
const AdminRole = "king"

// RBACMiddleware checks if the user has permission to access the resource
func RBACMiddleware(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, ok := userInterface.(ResourceChecker)
		if !ok {
			c.JSON(401, gin.H{"error": "invalid user data"})
			c.Abort()
			return
		}

		// Admin has access to everything
		if userInterface.(User).IsAdmin() {
			c.Next()
			return
		}

		// Check if user has permission for this resource
		if !user.CanAccessResource(resource) {
			c.JSON(403, gin.H{
				"error": "forbidden",
				"message": "you don't have permission to access this resource",
				"required_resource": resource,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// EndpointChecker checks if a user can access a specific endpoint
type EndpointChecker interface {
	CanAccessEndpoint(endpoint string) bool
}

// RBACEndpointMiddleware checks if the user can access the specific endpoint
func RBACEndpointMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, ok := userInterface.(EndpointChecker)
		if !ok {
			c.JSON(401, gin.H{"error": "invalid user data"})
			c.Abort()
			return
		}

		// Admin has access to all endpoints
		if userInterface.(User).IsAdmin() {
			c.Next()
			return
		}

		// Check endpoint-based access
		if !user.CanAccessEndpoint(c.Request.URL.Path) {
			c.JSON(403, gin.H{
				"error": "forbidden",
				"message": "you don't have permission to access this endpoint",
				"endpoint": c.Request.URL.Path,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser returns the current user from context
func GetCurrentUser(c *gin.Context) (User, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}

	user, ok := userInterface.(User)
	if !ok {
		return nil, errors.New("invalid user data")
	}

	return user, nil
}

// ExtractBearerToken extracts the bearer token from the Authorization header
func ExtractBearerToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
