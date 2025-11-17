// Package logging provides structured logging and metrics collection functionality.
package logging

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContextKey type for context keys
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
	metrics *Metrics
}

// NewLogger creates a new logger instance with metrics
func NewLogger(level string) (*Logger, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Parse and set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(logLevel)

	return &Logger{
		Logger:  logger,
		metrics: NewMetrics(),
	}, nil
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(ctx context.Context) *logrus.Entry {
	requestID := GetRequestID(ctx)
	return l.WithField("request_id", requestID)
}

// WithContext creates a logger entry with context information
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.WithRequestID(ctx)
	return entry
}

// GetMetrics returns the metrics collector
func (l *Logger) GetMetrics() *Metrics {
	return l.metrics
}

// GenerateRequestID generates a new unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}

// WithRequestIDContext adds a request ID to the context
func WithRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}
