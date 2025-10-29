package dragon

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up HTTP routes for dragon service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		dragons := api.Group("/dragons")
		{
			// Dragon CRUD operations
			dragons.POST("", handler.CreateDragon)                    // Create dragon
			dragons.POST("/:id/attack", handler.AttackDragon)         // Attack dragon
			dragons.GET("/:id", handler.GetDragon)                    // Get dragon by ID
			dragons.GET("/type/:type", handler.GetDragonsByType)      // Get dragons by type
			dragons.GET("/creator/:creator", handler.GetDragonsByCreator) // Get dragons by creator
		}
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "dragon",
		})
	})
}
