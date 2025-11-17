package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJQTransformer_TransformWithQuery(t *testing.T) {
	transformer := NewJQTransformer()

	// Sample data for testing
	sampleData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
				"profile": map[string]interface{}{
					"age":  30,
					"city": "New York",
				},
			},
			map[string]interface{}{
				"id":    2,
				"name":  "Jane Smith",
				"email": "jane@example.com",
				"profile": map[string]interface{}{
					"age":  25,
					"city": "Los Angeles",
				},
			},
		},
		"total": 2,
		"page":  1,
	}

	tests := []struct {
		name        string
		data        interface{}
		query       string
		expected    interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:     "simple field extraction",
			data:     sampleData,
			query:    ".total",
			expected: 2,
		},
		{
			name:     "array length",
			data:     sampleData,
			query:    ".users | length",
			expected: 2,
		},
		{
			name:     "map over array",
			data:     sampleData,
			query:    ".users | map(.name)",
			expected: []interface{}{"John Doe", "Jane Smith"},
		},
		{
			name:  "select and transform",
			data:  sampleData,
			query: ".users | map({id: .id, name: .name})",
			expected: []interface{}{
				map[string]interface{}{"id": 1, "name": "John Doe"},
				map[string]interface{}{"id": 2, "name": "Jane Smith"},
			},
		},
		{
			name:     "filter and map",
			data:     sampleData,
			query:    ".users | map(select(.profile.age > 25)) | map(.name)",
			expected: []interface{}{"John Doe"},
		},
		{
			name:     "nested field extraction",
			data:     sampleData,
			query:    ".users[0].profile.city",
			expected: "New York",
		},
		{
			name:  "complex transformation",
			data:  sampleData,
			query: "{total_users: .total, user_names: [.users[].name], avg_age: (.users | map(.profile.age) | add / length)}",
			expected: map[string]interface{}{
				"total_users": 2,
				"user_names":  []interface{}{"John Doe", "Jane Smith"},
				"avg_age":     27.5,
			},
		},
		{
			name:     "identity query",
			data:     sampleData,
			query:    ".",
			expected: sampleData,
		},
		{
			name:     "empty query",
			data:     sampleData,
			query:    "",
			expected: sampleData,
		},
		{
			name:        "invalid query",
			data:        sampleData,
			query:       ".users | invalid_function",
			expectError: true,
			errorMsg:    "failed to compile jq query",
		},
		{
			name:        "syntax error",
			data:        sampleData,
			query:       ".users | map(",
			expectError: true,
			errorMsg:    "invalid jq query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformWithQuery(tt.data, tt.query)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestJQTransformer_ValidateQuery(t *testing.T) {
	transformer := NewJQTransformer()

	tests := []struct {
		name        string
		query       string
		expectError bool
		errorMsg    string
	}{
		{
			name:  "valid simple query",
			query: ".name",
		},
		{
			name:  "valid complex query",
			query: ".users | map(select(.age > 25)) | map({name: .name, age: .age})",
		},
		{
			name:  "empty query",
			query: "",
		},
		{
			name:        "invalid syntax",
			query:       ".users | map(",
			expectError: true,
			errorMsg:    "invalid jq query",
		},
		{
			name:        "invalid function",
			query:       ".users | invalid_function()",
			expectError: true,
			errorMsg:    "invalid jq query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer.ValidateQuery(tt.query)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJQTransformer_Transform_Legacy(t *testing.T) {
	transformer := NewJQTransformer()

	// The TransformWithQuery method should work with jq queries
	result, err := transformer.TransformWithQuery(map[string]interface{}{"test": "value"}, "{result: .test}")

	assert.NoError(t, err)
	expected := map[string]interface{}{"result": "value"}
	assert.Equal(t, expected, result)
}

func TestJQTransformer_EdgeCases(t *testing.T) {
	transformer := NewJQTransformer()

	tests := []struct {
		name     string
		data     interface{}
		query    string
		expected interface{}
	}{
		{
			name:     "null data",
			data:     nil,
			query:    ".",
			expected: nil,
		},
		{
			name:     "empty object",
			data:     map[string]interface{}{},
			query:    ".nonexistent",
			expected: nil,
		},
		{
			name:     "empty array",
			data:     []interface{}{},
			query:    ".[0]",
			expected: nil,
		},
		{
			name:     "string data",
			data:     "hello world",
			query:    ". | length",
			expected: 11,
		},
		{
			name:     "number data",
			data:     42,
			query:    ". * 2",
			expected: 84,
		},
		{
			name:     "boolean data",
			data:     true,
			query:    ". and false",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformWithQuery(tt.data, tt.query)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
