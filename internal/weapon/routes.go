package weapon

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the weapon service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api")
	{
		// Protected routes
		protected := api.Group("")
		protected.Use(AuthMiddleware())
		{
			// Get all weapons
			protected.GET("/weapons", handler.GetWeapons)

			// Get my weapons
			protected.GET("/weapons/my-weapons", handler.GetMyWeapons)

			// Buy weapon
			protected.POST("/weapons/buy", handler.BuyWeapon)

			// Admin routes (Light Emperor/King only)
			protected.POST("/weapons", handler.CreateWeapon)
		}
	}
}
