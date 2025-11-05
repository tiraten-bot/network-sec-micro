package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy  Status = "unhealthy"
	StatusDegraded   Status = "degraded"
)

// Component represents a health check component
type Component struct {
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string      `json:"status"`
	Timestamp  time.Time   `json:"timestamp"`
	Components []Component `json:"components,omitempty"`
}

// Checker defines the interface for health checks
type Checker interface {
	Check() Component
}

// Handler handles health check requests
type Handler struct {
	checkers []Checker
}

// NewHandler creates a new health check handler
func NewHandler(checkers ...Checker) *Handler {
	return &Handler{
		checkers: checkers,
	}
}

// RegisterChecker registers a new health checker
func (h *Handler) RegisterChecker(checker Checker) {
	h.checkers = append(h.checkers, checker)
}

// Health returns the health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	components := make([]Component, 0, len(h.checkers))
	overallStatus := StatusHealthy

	for _, checker := range h.checkers {
		component := checker.Check()
		components = append(components, component)

		if component.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if component.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	response := HealthResponse{
		Status:     string(overallStatus),
		Timestamp:  time.Now(),
		Components: components,
	}

	w.Header().Set("Content-Type", "application/json")
	
	if overallStatus == StatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if overallStatus == StatusDegraded {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

// Ready returns the readiness status (simpler than health)
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Live returns the liveness status
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

