package health

import (
	"time"

	"gorm.io/gorm"
)

// DatabaseChecker checks database health
type DatabaseChecker struct {
	DB     *gorm.DB
	DBName string
}

// Check performs a database health check
func (c *DatabaseChecker) Check() Component {
	component := Component{
		Name:      c.DBName,
		Timestamp: time.Now(),
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		component.Status = StatusUnhealthy
		component.Message = "Failed to get database connection: " + err.Error()
		return component
	}

	if err := sqlDB.Ping(); err != nil {
		component.Status = StatusUnhealthy
		component.Message = "Database ping failed: " + err.Error()
		return component
	}

	stats := sqlDB.Stats()
	if stats.OpenConnections > stats.MaxOpenConnections*80/100 {
		component.Status = StatusDegraded
		component.Message = "High connection usage"
		return component
	}

	component.Status = StatusHealthy
	component.Message = "Database connection healthy"
	return component
}

