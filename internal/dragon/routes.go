package dragon

import (
	"net/http"
	"time"

	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes sets up HTTP routes for dragon service
func SetupRoutes(r *gin.Engine, handler *Handler) {
	api := r.Group("/api/v1")
	{
		dragons := api.Group("/dragons")
		{
			// Dragon CRUD operations
			dragons.POST("", handler.CreateDragon)                    // Create dragon
			dragons.POST("/:id/attack", handler.AttackDragon)         // Attack dragon
			dragons.POST("/:id/revive", handler.ReviveDragon)         // Revive dragon (for battle service)
			dragons.GET("/:id", handler.GetDragon)                    // Get dragon by ID
			dragons.GET("/type/:type", handler.GetDragonsByType)      // Get dragons by type
			dragons.GET("/creator/:creator", handler.GetDragonsByCreator) // Get dragons by creator
		}
	}

	// Health check endpoints
	healthHandler := health.NewHandler(&health.MongoDBChecker{Client: Client, DBName: "mongodb"})
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
}
