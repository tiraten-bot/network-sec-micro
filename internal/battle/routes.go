package battle

import (
	"net/http"
	"time"

	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes configures all routes for the battle service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	// Health check endpoints
	var dbChecker health.Checker
	if SQLDB.Enabled {
		if gormDB, ok := SQLDB.DB.(*gorm.DB); ok {
			dbChecker = &health.DatabaseChecker{DB: gormDB, DBName: "postgres"}
		}
	}
	if dbChecker == nil {
		// Fallback: create a simple checker that always returns healthy if SQL not enabled
		dbChecker = &health.SimpleChecker{Name: "postgres", Status: health.StatusHealthy, Message: "PostgreSQL not enabled"}
	}
	healthHandler := health.NewHandler(dbChecker)
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
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()
		statusText := http.StatusText(status)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusText).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path, statusText).Observe(duration)
	})

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

