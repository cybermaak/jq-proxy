package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"jq-proxy-service/internal/client"
	"jq-proxy-service/internal/logging"
	"jq-proxy-service/internal/models"
	"jq-proxy-service/internal/proxy"
	"jq-proxy-service/internal/transform"
)

// BenchmarkSuite contains performance benchmarks
type BenchmarkSuite struct {
	mockServer   *httptest.Server
	proxyHandler http.Handler
	sampleData   map[string]interface{}
}

// setupBenchmark initializes the benchmark environment
func setupBenchmark() *BenchmarkSuite {
	suite := &BenchmarkSuite{}

	// Sample data for consistent benchmarking
	suite.sampleData = map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
				"profile": map[string]interface{}{
					"age":     30,
					"city":    "New York",
					"country": "USA",
				},
			},
			map[string]interface{}{
				"id":    2,
				"name":  "Jane Smith",
				"email": "jane@example.com",
				"profile": map[string]interface{}{
					"age":     25,
					"city":    "Los Angeles",
					"country": "USA",
				},
			},
		},
		"total": 2,
		"page":  1,
	}

	// Create mock server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(suite.sampleData)
	}))

	// Setup proxy service
	configData := fmt.Sprintf(`{
		"server": {"port": 8080, "read_timeout": 30, "write_timeout": 30},
		"endpoints": {
			"test-api": {"name": "test-api", "target": "%s"}
		}
	}`, suite.mockServer.URL)

	configProvider := &mockConfigProvider{configData: configData}
	httpClient := client.NewClient(30 * time.Second)
	transformer := transform.NewUnifiedTransformer()

	logger, _ := logging.NewLogger("error")

	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)
	handler := proxy.NewHandler(proxyService, logger)

	suite.proxyHandler = handler.SetupRoutes()
	return suite
}

// BenchmarkJQTransformation benchmarks jq transformations
func BenchmarkJQTransformation(b *testing.B) {
	suite := setupBenchmark()
	defer suite.mockServer.Close()

	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{user_count: .total, user_names: [.data[].name], user_ages: [.data[].profile.age]}",
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.proxyHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rr.Code)
		}
	}
}

// BenchmarkJQTransformationHTTP benchmarks jq transformations via HTTP
func BenchmarkJQTransformationHTTP(b *testing.B) {
	suite := setupBenchmark()
	defer suite.mockServer.Close()

	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{user_count: .total, user_names: [.data[].name], user_ages: [.data[].profile.age]}",
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.proxyHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rr.Code)
		}
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	suite := setupBenchmark()
	defer suite.mockServer.Close()

	requestBody := map[string]interface{}{
		"method": "GET",
		"body":   nil,
		"transformation": map[string]interface{}{
			"result": "$.data[*].name",
		},
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			suite.proxyHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", rr.Code)
			}
		}
	})
}

// BenchmarkTransformationOnly benchmarks just the transformation logic
func BenchmarkTransformationOnly(b *testing.B) {
	transformer := transform.NewUnifiedTransformer()

	sampleData := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": 1, "name": "John", "age": 30},
			map[string]interface{}{"id": 2, "name": "Jane", "age": 25},
		},
		"total": 2,
	}

	b.Run("jq", func(b *testing.B) {
		req := &models.ProxyRequest{
			Method:             "GET",
			TransformationMode: models.TransformationModeJQ,
			JQQuery:            "{names: [.data[].name], count: .total}",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err := transformer.TransformRequest(sampleData, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage under load
func BenchmarkMemoryUsage(b *testing.B) {
	suite := setupBenchmark()
	defer suite.mockServer.Close()

	// Create a larger dataset
	largeData := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]interface{}{
			"id":   i,
			"name": fmt.Sprintf("User %d", i),
			"data": map[string]interface{}{
				"value": i * 2,
				"text":  fmt.Sprintf("Text for user %d", i),
			},
		}
	}

	// Update mock server to return large dataset
	suite.mockServer.Close()
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":  largeData,
			"total": len(largeData),
		})
	}))

	// Update config
	configData := fmt.Sprintf(`{
		"server": {"port": 8080, "read_timeout": 30, "write_timeout": 30},
		"endpoints": {
			"test-api": {"name": "test-api", "target": "%s"}
		}
	}`, suite.mockServer.URL)

	configProvider := &mockConfigProvider{configData: configData}
	httpClient := client.NewClient(30 * time.Second)
	transformer := transform.NewUnifiedTransformer()

	logger, _ := logging.NewLogger("error")

	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)
	handler := proxy.NewHandler(proxyService, logger)
	suite.proxyHandler = handler.SetupRoutes()

	requestBody := map[string]interface{}{
		"method": "GET",
		"body":   nil,
		"transformation": map[string]interface{}{
			"user_names": "$.data[*].name",
			"user_count": "$.total",
		},
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		suite.proxyHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", rr.Code)
		}
	}
}

// BenchmarkHighConcurrency tests performance under high concurrency
func BenchmarkHighConcurrency(b *testing.B) {
	suite := setupBenchmark()
	defer suite.mockServer.Close()

	requestBody := map[string]interface{}{
		"method": "GET",
		"body":   nil,
		"transformation": map[string]interface{}{
			"result": "$.data[*].name",
		},
	}

	body, _ := json.Marshal(requestBody)
	concurrency := 100

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(concurrency)

		for j := 0; j < concurrency; j++ {
			go func() {
				defer wg.Done()

				req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				rr := httptest.NewRecorder()
				suite.proxyHandler.ServeHTTP(rr, req)

				if rr.Code != http.StatusOK {
					b.Errorf("Expected status 200, got %d", rr.Code)
				}
			}()
		}

		wg.Wait()
	}
}

// mockConfigProvider for benchmarks
type mockConfigProvider struct {
	configData string
}

func (m *mockConfigProvider) LoadConfig() (*models.ProxyConfig, error) {
	return models.ParseProxyConfig([]byte(m.configData))
}

func (m *mockConfigProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, false
	}
	endpoint, exists := config.Endpoints[name]
	return endpoint, exists
}

func (m *mockConfigProvider) Reload() error {
	_, err := m.LoadConfig()
	return err
}
