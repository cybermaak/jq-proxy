package logging

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if metrics.endpointMetrics == nil {
		t.Error("Expected initialized endpoint metrics map")
	}
}

func TestRecordRequest(t *testing.T) {
	metrics := NewMetrics()
	endpoint := "test-endpoint"
	duration := 100 * time.Millisecond

	metrics.RecordRequest(endpoint, duration)

	snapshot := metrics.GetMetrics()

	if snapshot.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", snapshot.TotalRequests)
	}

	if snapshot.AverageResponseTime != duration {
		t.Errorf("Expected average response time %v, got %v", duration, snapshot.AverageResponseTime)
	}

	em, exists := snapshot.Endpoints[endpoint]
	if !exists {
		t.Fatal("Expected endpoint metrics to exist")
	}

	if em.RequestCount != 1 {
		t.Errorf("Expected 1 request for endpoint, got %d", em.RequestCount)
	}

	if em.AvgResponseTime != duration {
		t.Errorf("Expected avg response time %v, got %v", duration, em.AvgResponseTime)
	}
}

func TestRecordMultipleRequests(t *testing.T) {
	metrics := NewMetrics()
	endpoint := "test-endpoint"

	metrics.RecordRequest(endpoint, 100*time.Millisecond)
	metrics.RecordRequest(endpoint, 200*time.Millisecond)

	snapshot := metrics.GetMetrics()

	if snapshot.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", snapshot.TotalRequests)
	}

	expectedAvg := 150 * time.Millisecond
	if snapshot.AverageResponseTime != expectedAvg {
		t.Errorf("Expected average response time %v, got %v", expectedAvg, snapshot.AverageResponseTime)
	}

	em := snapshot.Endpoints[endpoint]
	if em.RequestCount != 2 {
		t.Errorf("Expected 2 requests for endpoint, got %d", em.RequestCount)
	}

	if em.AvgResponseTime != expectedAvg {
		t.Errorf("Expected endpoint avg response time %v, got %v", expectedAvg, em.AvgResponseTime)
	}
}

func TestRecordError(t *testing.T) {
	metrics := NewMetrics()
	endpoint := "test-endpoint"

	metrics.RecordError(endpoint)

	snapshot := metrics.GetMetrics()

	if snapshot.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", snapshot.TotalErrors)
	}

	em, exists := snapshot.Endpoints[endpoint]
	if !exists {
		t.Fatal("Expected endpoint metrics to exist")
	}

	if em.ErrorCount != 1 {
		t.Errorf("Expected 1 error for endpoint, got %d", em.ErrorCount)
	}
}

func TestRecordMultipleEndpoints(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordRequest("endpoint1", 100*time.Millisecond)
	metrics.RecordRequest("endpoint2", 200*time.Millisecond)
	metrics.RecordError("endpoint1")

	snapshot := metrics.GetMetrics()

	if snapshot.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", snapshot.TotalRequests)
	}

	if snapshot.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", snapshot.TotalErrors)
	}

	if len(snapshot.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(snapshot.Endpoints))
	}

	em1 := snapshot.Endpoints["endpoint1"]
	if em1.RequestCount != 1 {
		t.Errorf("Expected 1 request for endpoint1, got %d", em1.RequestCount)
	}
	if em1.ErrorCount != 1 {
		t.Errorf("Expected 1 error for endpoint1, got %d", em1.ErrorCount)
	}

	em2 := snapshot.Endpoints["endpoint2"]
	if em2.RequestCount != 1 {
		t.Errorf("Expected 1 request for endpoint2, got %d", em2.RequestCount)
	}
	if em2.ErrorCount != 0 {
		t.Errorf("Expected 0 errors for endpoint2, got %d", em2.ErrorCount)
	}
}

func TestGetMetricsSnapshot(t *testing.T) {
	metrics := NewMetrics()

	// Record some data
	metrics.RecordRequest("endpoint1", 100*time.Millisecond)
	metrics.RecordError("endpoint1")

	// Get first snapshot
	snapshot1 := metrics.GetMetrics()

	// Record more data
	metrics.RecordRequest("endpoint1", 200*time.Millisecond)

	// Get second snapshot
	snapshot2 := metrics.GetMetrics()

	// Verify snapshots are independent
	if snapshot1.TotalRequests == snapshot2.TotalRequests {
		t.Error("Expected snapshots to have different request counts")
	}

	if snapshot1.TotalRequests != 1 {
		t.Errorf("Expected first snapshot to have 1 request, got %d", snapshot1.TotalRequests)
	}

	if snapshot2.TotalRequests != 2 {
		t.Errorf("Expected second snapshot to have 2 requests, got %d", snapshot2.TotalRequests)
	}
}
