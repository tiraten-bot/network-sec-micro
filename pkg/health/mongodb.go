package health

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBChecker checks MongoDB health
type MongoDBChecker struct {
	Client *mongo.Client
	DBName string
}

// Check performs a MongoDB health check
func (c *MongoDBChecker) Check() Component {
	component := Component{
		Name:      c.DBName,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Client.Ping(ctx, nil); err != nil {
		component.Status = StatusUnhealthy
		component.Message = "MongoDB ping failed: " + err.Error()
		return component
	}

	component.Status = StatusHealthy
	component.Message = "MongoDB connection healthy"
	return component
}

