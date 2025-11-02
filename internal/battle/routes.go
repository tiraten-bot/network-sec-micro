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
			protected.GET("/battles/:id", handler.GetBattle)
			protected.GET("/battles/my-battles", handler.GetMyBattles)
			protected.GET("/battles/stats", handler.GetBattleStats)
			protected.GET("/battles/:id/turns", handler.GetBattleTurns)

			// Battle actions
			protected.POST("/battles/attack", handler.Attack)
		}
	}
}

