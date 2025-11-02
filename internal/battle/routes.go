package battle

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the battle service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api")
	{
		// Protected routes
		protected := api.Group("")
		protected.Use(AuthMiddleware())
		{
			// Battle CRUD operations
			protected.POST("/battles", handler.StartBattle)
			protected.POST("/battles/attack", handler.Attack)

			// RBAC protected routes
			rbac := protected.Group("")
			rbac.Use(RBACMiddleware())
			{
				rbac.GET("/battles/:id", handler.GetBattle)
				rbac.GET("/battles/my-battles", handler.GetMyBattles)
				rbac.GET("/battles/stats", handler.GetBattleStats)
				rbac.GET("/battles/:id/turns", handler.GetBattleTurns)
			}
		}
	}
}

