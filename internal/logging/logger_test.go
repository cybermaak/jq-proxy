package logging

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel logrus.Level
		wantErr   bool
	}{
		{
			name:      "debug level",
			level:     "debug",
			wantLevel: logrus.DebugLevel,
			wantErr:   false,
		},
		{
			name:      "info level",
			level:     "info",
			wantLevel: logrus.InfoLevel,
			wantErr:   false,
		},
		{
			name:      "warn level",
			level:     "warn",
			wantLevel: logrus.WarnLevel,
			wantErr:   false,
		},
		{
			name:      "error level",
			level:     "error",
			wantLevel: logrus.ErrorLevel,
			wantErr:   false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.level)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if logger.Logger.Level != tt.wantLevel {
				t.Errorf("Expected level %v, got %v", tt.wantLevel, logger.Logger.Level)
			}

			if logger.metrics == nil {
				t.Error("Expected metrics to be initialized")
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("Expected non-empty request ID")
	}

	if id1 == id2 {
		t.Error("Expected unique request IDs")
	}
}

func TestWithRequestIDContext(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-id"

	ctx = WithRequestIDContext(ctx, requestID)

	retrievedID := GetRequestID(ctx)
	if retrievedID != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, retrievedID)
	}
}

func TestGetRequestIDWithoutContext(t *testing.T) {
	ctx := context.Background()
	requestID := GetRequestID(ctx)

	if requestID != "unknown" {
		t.Errorf("Expected 'unknown', got %s", requestID)
	}
}

func TestWithRequestID(t *testing.T) {
	logger, err := NewLogger("info")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	requestID := "test-request-id"
	ctx = WithRequestIDContext(ctx, requestID)

	entry := logger.WithRequestID(ctx)

	if entry.Data["request_id"] != requestID {
		t.Errorf("Expected request_id %s in entry, got %v", requestID, entry.Data["request_id"])
	}
}

func TestWithContext(t *testing.T) {
	logger, err := NewLogger("info")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	requestID := "test-request-id"
	ctx = WithRequestIDContext(ctx, requestID)

	entry := logger.WithContext(ctx)

	if entry.Data["request_id"] != requestID {
		t.Errorf("Expected request_id %s in entry, got %v", requestID, entry.Data["request_id"])
	}
}

func TestGetMetrics(t *testing.T) {
	logger, err := NewLogger("info")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := logger.GetMetrics()
	if metrics == nil {
		t.Error("Expected non-nil metrics")
	}
}
