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
			protected.POST("/battles/revive-dragon", handler.ReviveDragon)
			protected.POST("/battles/dark-emperor-join", handler.DarkEmperorJoinBattle)
			protected.POST("/battles/sacrifice-dragon", handler.SacrificeDragon)
			// Spell casting moved to battlespell service - endpoint removed
			
			// Arena battle endpoint (called by arena service)
			api.POST("/arena/start", handler.StartArenaBattle)

			// RBAC protected routes
			rbac := protected.Group("")
			rbac.Use(RBACMiddleware())
			{
			rbac.GET("/battles/:id", handler.GetBattle)
			rbac.GET("/battles/my-battles", handler.GetMyBattles)
			rbac.GET("/battles/stats", handler.GetBattleStats)
			rbac.GET("/battles/:id/turns", handler.GetBattleTurns)
			rbac.GET("/battles/:id/logs", handler.GetBattleLogs)
			}
		}
	}
}

