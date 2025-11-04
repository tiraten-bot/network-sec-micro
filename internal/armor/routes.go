package armor

import "github.com/gin-gonic/gin"

// SetupRoutes configures all routes for the armor service
func SetupRoutes(r *gin.Engine, handler *Handler) {
    api := r.Group("/api")
    {
        protected := api.Group("")
        protected.Use(AuthMiddleware())
        {
            protected.GET("/armors", handler.GetArmors)
            protected.GET("/armors/my-armors", handler.GetMyArmors)
            protected.POST("/armors/buy", handler.BuyArmor)
            protected.POST("/armors", handler.CreateArmor)
        }
    }
}


