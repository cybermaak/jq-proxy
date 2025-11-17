// Package logging provides structured logging and metrics collection functionality.
package logging

import (
	"sync"
	"time"
)

// Metrics collects application metrics
type Metrics struct {
	mu                sync.RWMutex
	requestCount      int64
	errorCount        int64
	totalResponseTime time.Duration
	endpointMetrics   map[string]*EndpointMetrics
}

// EndpointMetrics tracks metrics for a specific endpoint
type EndpointMetrics struct {
	RequestCount      int64
	ErrorCount        int64
	TotalResponseTime time.Duration
	AvgResponseTime   time.Duration
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		endpointMetrics: make(map[string]*EndpointMetrics),
	}
}

// RecordRequest records a successful request
func (m *Metrics) RecordRequest(endpoint string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount++
	m.totalResponseTime += duration

	if _, exists := m.endpointMetrics[endpoint]; !exists {
		m.endpointMetrics[endpoint] = &EndpointMetrics{}
	}

	em := m.endpointMetrics[endpoint]
	em.RequestCount++
	em.TotalResponseTime += duration
	em.AvgResponseTime = time.Duration(int64(em.TotalResponseTime) / em.RequestCount)
}

// RecordError records a failed request
func (m *Metrics) RecordError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorCount++

	if _, exists := m.endpointMetrics[endpoint]; !exists {
		m.endpointMetrics[endpoint] = &EndpointMetrics{}
	}

	m.endpointMetrics[endpoint].ErrorCount++
}

// GetMetrics returns a snapshot of current metrics
func (m *Metrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgResponseTime := time.Duration(0)
	if m.requestCount > 0 {
		avgResponseTime = time.Duration(int64(m.totalResponseTime) / m.requestCount)
	}

	endpoints := make(map[string]EndpointMetrics)
	for name, em := range m.endpointMetrics {
		endpoints[name] = *em
	}

	return MetricsSnapshot{
		TotalRequests:       m.requestCount,
		TotalErrors:         m.errorCount,
		AverageResponseTime: avgResponseTime,
		Endpoints:           endpoints,
	}
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	TotalRequests       int64                      `json:"total_requests"`
	TotalErrors         int64                      `json:"total_errors"`
	AverageResponseTime time.Duration              `json:"average_response_time"`
	Endpoints           map[string]EndpointMetrics `json:"endpoints"`
}
