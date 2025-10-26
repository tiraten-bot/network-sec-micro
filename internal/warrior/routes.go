package warrior

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the warrior service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/login", handler.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(AuthMiddleware())
		{
			// Profile route (accessible by all authenticated users)
			protected.GET("/profile", handler.GetProfile)

			// Admin routes (King only)
			protected.GET("/warriors", handler.GetWarriors)

			// Knight endpoints
			knight := protected.Group("")
			knight.Use(RBACEndpointMiddleware())
			{
				knight.GET("/weapons", handler.GetWeapons)
				knight.GET("/armor", handler.GetArmor)
				knight.GET("/battles", handler.GetBattles)
			}

			// Archer endpoints
			archer := protected.Group("")
			archer.Use(RBACEndpointMiddleware())
			{
				archer.GET("/arrows", handler.GetArrows)
				archer.GET("/scouting", handler.GetScouting)
			}

			// Mage endpoints
			mage := protected.Group("")
			mage.Use(RBACEndpointMiddleware())
			{
				mage.GET("/spells", handler.GetSpells)
				mage.GET("/potions", handler.GetPotions)
				mage.GET("/library", handler.GetLibrary)
			}
		}
	}
}
