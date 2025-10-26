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

			// Knight endpoints (accessible by Knight and King)
			knight := protected.Group("")
			knight.Use(RBACEndpointMiddleware())
			{
				knight.GET("/warriors/knights", handler.GetKnightWarriors)
			}

			// Archer endpoints (accessible by Archer and King)
			archer := protected.Group("")
			archer.Use(RBACEndpointMiddleware())
			{
				archer.GET("/warriors/archers", handler.GetArcherWarriors)
			}

			// Mage endpoints (accessible by Mage and King)
			mage := protected.Group("")
			mage.Use(RBACEndpointMiddleware())
			{
				mage.GET("/warriors/mages", handler.GetMageWarriors)
			}
		}
	}
}