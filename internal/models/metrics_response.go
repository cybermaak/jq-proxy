package models

const MetricsSchemaVersion = "1.0"

type MetricsResponse struct {
	SchemaVersion         string                             `json:"schema_version"`
	TotalRequests         int64                              `json:"total_requests"`
	TotalErrors           int64                              `json:"total_errors"`
	AverageResponseTimeNs int64                              `json:"average_response_time_ns"`
	Endpoints             map[string]EndpointMetricsResponse `json:"endpoints"`
}

type EndpointMetricsResponse struct {
	RequestCount          int64 `json:"request_count"`
	ErrorCount            int64 `json:"error_count"`
	TotalResponseTimeNs   int64 `json:"total_response_time_ns"`
	AverageResponseTimeNs int64 `json:"average_response_time_ns"`
}
