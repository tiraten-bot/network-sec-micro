package arenaspell

import (
    "github.com/gin-gonic/gin"
)

// SetupRoutes configures HTTP routes for arenaspell service
func SetupRoutes(r *gin.Engine, handler *Handler) {
    api := r.Group("/api/v1")
    {
        api.POST("/arenaspells/cast", handler.CastSpell)
    }
}


