package health

import "time"

// SimpleChecker is a simple health checker that always returns a fixed status
type SimpleChecker struct {
	Name    string
	Status  Status
	Message string
}

// Check returns the component status
func (c *SimpleChecker) Check() Component {
	return Component{
		Name:      c.Name,
		Status:    c.Status,
		Message:   c.Message,
		Timestamp: time.Now(),
	}
}

