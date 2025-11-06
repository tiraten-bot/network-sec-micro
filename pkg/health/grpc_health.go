package health

import (
	"context"
	"time"
)

// GRPCHealthChecker is a simple checker for gRPC services
type GRPCHealthChecker struct {
	Name    string
	Status  Status
	Message string
}

// Check returns the component status
func (c *GRPCHealthChecker) Check() Component {
	return Component{
		Name:      c.Name,
		Status:    c.Status,
		Message:   c.Message,
		Timestamp: time.Now(),
	}
}

