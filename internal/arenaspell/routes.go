package arenaspell

import (
	"net/http"
	"time"

	"network-sec-micro/pkg/health"
	"network-sec-micro/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes configures HTTP routes for arenaspell service
func SetupRoutes(r *gin.Engine, handler *Handler) {
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

	api := r.Group("/api/v1")
    {
        api.POST("/arenaspells/cast", handler.CastSpell)
    }
}


