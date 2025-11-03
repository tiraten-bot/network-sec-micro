package arena

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures HTTP routes for arena service
func SetupRoutes(r *gin.Engine, handler *Handler) {
    api := r.Group("/api/v1/arena")
    api.Use(AuthMiddleware())
	{
		// Invitation operations
		api.POST("/invitations", handler.SendInvitation)
		api.POST("/invitations/accept", handler.AcceptInvitation)
		api.POST("/invitations/reject", handler.RejectInvitation)
		api.POST("/invitations/cancel", handler.CancelInvitation)
		api.GET("/invitations/:id", handler.GetInvitation)
		api.GET("/invitations/my", handler.GetMyInvitations)

		// Match operations
		api.GET("/matches/my", handler.GetMyMatches)
		api.GET("/matches/:id", handler.GetMatch)
		api.POST("/matches/attack", handler.AttackInArena)
		// Arenaspell application
		api.POST("/spells/apply", handler.ApplyArenaSpell)
	}
}

