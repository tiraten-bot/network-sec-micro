package warrior

import (
	"net/http"

	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes configures all routes for the warrior service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	// Health check endpoints
	healthHandler := health.NewHandler(&health.DatabaseChecker{DB: DB, DBName: "postgres"})
	r.GET("/health", func(c *gin.Context) {
		healthHandler.Health(c.Writer, c.Request)
	})
	r.GET("/ready", func(c *gin.Context) {
		healthHandler.Ready(c.Writer, c.Request)
	})
	r.GET("/live", func(c *gin.Context) {
		healthHandler.Live(c.Writer, c.Request)
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Metrics middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path, http.StatusText(status)).Observe(duration)
	})

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
			protected.GET("/profile/kills", handler.GetMyKilledMonsters)
			protected.GET("/profile/strongest-kill", handler.GetMyStrongestKill)

			// Password management
			protected.PUT("/profile/password", handler.ChangePassword)

			// Admin routes (King only)
			protected.POST("/warriors", handler.CreateWarrior)
			protected.GET("/warriors", handler.GetWarriors)

			// Individual warrior routes
			protected.GET("/warriors/:id", handler.GetWarriorById)
			protected.PUT("/warriors/:id", handler.UpdateWarrior)
			protected.DELETE("/warriors/:id", handler.DeleteWarrior)

			// Role-based warrior endpoints
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