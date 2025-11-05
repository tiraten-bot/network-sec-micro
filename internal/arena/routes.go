package arena

import (
	"net/http"
	"time"

	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

// SetupRoutes configures HTTP routes for arena service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	// Health check endpoints
	var dbChecker health.Checker
	if SQLDB.Enabled {
		if gormDB, ok := SQLDB.DB.(*gorm.DB); ok {
			dbChecker = &health.DatabaseChecker{DB: gormDB, DBName: "postgres"}
		}
	}
	if dbChecker == nil {
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

