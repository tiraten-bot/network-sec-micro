package battlespell

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures HTTP routes for battlespell service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		// Spell operations
		api.POST("/spells/cast", handler.CastSpell)
	}
}

