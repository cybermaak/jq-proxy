package metrics

import (
	"sync"
	"time"
)

// Metrics collects and tracks service metrics
type Metrics struct {
	mu                    sync.RWMutex
	requestCount          int64
	errorCount            int64
	transformationCount   int64
	endpointRequestCounts map[string]int64
	responseTimes         []time.Duration
	transformationModes   map[string]int64
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		endpointRequestCounts: make(map[string]int64),
		responseTimes:         make([]time.Duration, 0),
		transformationModes:   make(map[string]int64),
	}
}

// IncrementRequestCount increments the total request count
func (m *Metrics) IncrementRequestCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount++
}

// IncrementErrorCount increments the error count
func (m *Metrics) IncrementErrorCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCount++
}

// IncrementTransformationCount increments the transformation count
func (m *Metrics) IncrementTransformationCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transformationCount++
}

// IncrementEndpointRequestCount increments the request count for a specific endpoint
func (m *Metrics) IncrementEndpointRequestCount(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endpointRequestCounts[endpoint]++
}

// RecordResponseTime records a response time
func (m *Metrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Keep only the last 1000 response times to prevent memory growth
	if len(m.responseTimes) >= 1000 {
		m.responseTimes = m.responseTimes[1:]
	}
	m.responseTimes = append(m.responseTimes, duration)
}

// IncrementTransformationMode increments the count for a transformation mode
func (m *Metrics) IncrementTransformationMode(mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transformationModes[mode]++
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		RequestCount:          m.requestCount,
		ErrorCount:            m.errorCount,
		TransformationCount:   m.transformationCount,
		EndpointRequestCounts: make(map[string]int64),
		TransformationModes:   make(map[string]int64),
	}

	// Copy endpoint counts
	for endpoint, count := range m.endpointRequestCounts {
		stats.EndpointRequestCounts[endpoint] = count
	}

	// Copy transformation mode counts
	for mode, count := range m.transformationModes {
		stats.TransformationModes[mode] = count
	}

	// Calculate response time statistics
	if len(m.responseTimes) > 0 {
		var total time.Duration
		min := m.responseTimes[0]
		max := m.responseTimes[0]

		for _, rt := range m.responseTimes {
			total += rt
			if rt < min {
				min = rt
			}
			if rt > max {
				max = rt
			}
		}

		stats.ResponseTimeStats = ResponseTimeStats{
			Count:   int64(len(m.responseTimes)),
			Average: total / time.Duration(len(m.responseTimes)),
			Min:     min,
			Max:     max,
		}
	}

	return stats
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCount = 0
	m.errorCount = 0
	m.transformationCount = 0
	m.endpointRequestCounts = make(map[string]int64)
	m.responseTimes = make([]time.Duration, 0)
	m.transformationModes = make(map[string]int64)
}

// Stats represents collected metrics statistics
type Stats struct {
	RequestCount          int64             `json:"request_count"`
	ErrorCount            int64             `json:"error_count"`
	TransformationCount   int64             `json:"transformation_count"`
	EndpointRequestCounts map[string]int64  `json:"endpoint_request_counts"`
	TransformationModes   map[string]int64  `json:"transformation_modes"`
	ResponseTimeStats     ResponseTimeStats `json:"response_time_stats"`
}

// ResponseTimeStats contains response time statistics
type ResponseTimeStats struct {
	Count   int64         `json:"count"`
	Average time.Duration `json:"average"`
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
}

// ErrorRate calculates the error rate as a percentage
func (s *Stats) ErrorRate() float64 {
	if s.RequestCount == 0 {
		return 0
	}
	return float64(s.ErrorCount) / float64(s.RequestCount) * 100
}

// MostUsedEndpoint returns the endpoint with the most requests
func (s *Stats) MostUsedEndpoint() (string, int64) {
	var maxEndpoint string
	var maxCount int64

	for endpoint, count := range s.EndpointRequestCounts {
		if count > maxCount {
			maxCount = count
			maxEndpoint = endpoint
		}
	}

	return maxEndpoint, maxCount
}

// MostUsedTransformationMode returns the most used transformation mode
func (s *Stats) MostUsedTransformationMode() (string, int64) {
	var maxMode string
	var maxCount int64

	for mode, count := range s.TransformationModes {
		if count > maxCount {
			maxCount = count
			maxMode = mode
		}
	}

	return maxMode, maxCount
}
