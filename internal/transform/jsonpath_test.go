package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONPathTransformer_Transform(t *testing.T) {
	transformer := NewJSONPathTransformer()

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
		name           string
		data           interface{}
		transformation map[string]interface{}
		expected       map[string]interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "simple field extraction",
			data: sampleData,
			transformation: map[string]interface{}{
				"total_users":  "$.total",
				"current_page": "$.page",
			},
			expected: map[string]interface{}{
				"total_users":  2,
				"current_page": 1,
			},
		},
		{
			name: "array element extraction",
			data: sampleData,
			transformation: map[string]interface{}{
				"first_user_name": "$.users[0].name",
				"second_user_id":  "$.users[1].id",
			},
			expected: map[string]interface{}{
				"first_user_name": "John Doe",
				"second_user_id":  2,
			},
		},
		{
			name: "nested field extraction",
			data: sampleData,
			transformation: map[string]interface{}{
				"first_user_age":   "$.users[0].profile.age",
				"second_user_city": "$.users[1].profile.city",
			},
			expected: map[string]interface{}{
				"first_user_age":   30,
				"second_user_city": "Los Angeles",
			},
		},
		{
			name: "array mapping",
			data: sampleData,
			transformation: map[string]interface{}{
				"user_names": "$.users[*].name",
				"user_ids":   "$.users[*].id",
			},
			expected: map[string]interface{}{
				"user_names": []interface{}{"John Doe", "Jane Smith"},
				"user_ids":   []interface{}{1, 2},
			},
		},
		{
			name: "root path",
			data: sampleData,
			transformation: map[string]interface{}{
				"full_data": "$",
			},
			expected: map[string]interface{}{
				"full_data": sampleData,
			},
		},
		{
			name: "non-existent path",
			data: sampleData,
			transformation: map[string]interface{}{
				"missing_field": "$.nonexistent",
			},
			expected: map[string]interface{}{
				"missing_field": nil,
			},
		},
		{
			name:           "empty transformation",
			data:           sampleData,
			transformation: map[string]interface{}{},
			expected:       map[string]interface{}{},
		},
		{
			name:           "nil transformation",
			data:           sampleData,
			transformation: nil,
			expected:       sampleData,
		},
		{
			name: "invalid JSONPath expression",
			data: sampleData,
			transformation: map[string]interface{}{
				"invalid": "$.users[",
			},
			expectError: true,
			errorMsg:    "invalid JSONPath expression",
		},
		{
			name: "non-string transformation path",
			data: sampleData,
			transformation: map[string]interface{}{
				"invalid": 123,
			},
			expectError: true,
			errorMsg:    "transformation path for key 'invalid' must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.Transform(tt.data, tt.transformation)

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

func TestJSONPathTransformer_TransformWithFallback(t *testing.T) {
	transformer := &JSONPathTransformer{}

	sampleData := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	tests := []struct {
		name           string
		data           interface{}
		transformation map[string]interface{}
		fallbacks      map[string]interface{}
		expected       map[string]interface{}
	}{
		{
			name: "with fallback for missing field",
			data: sampleData,
			transformation: map[string]interface{}{
				"name":  "$.name",
				"email": "$.email", // missing field
			},
			fallbacks: map[string]interface{}{
				"email": "unknown@example.com",
			},
			expected: map[string]interface{}{
				"name":  "John",
				"email": "unknown@example.com",
			},
		},
		{
			name: "no fallback needed",
			data: sampleData,
			transformation: map[string]interface{}{
				"name": "$.name",
				"age":  "$.age",
			},
			fallbacks: map[string]interface{}{
				"email": "unknown@example.com",
			},
			expected: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformWithFallback(tt.data, tt.transformation, tt.fallbacks)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONPathTransformer_ValidateTransformation(t *testing.T) {
	transformer := &JSONPathTransformer{}

	tests := []struct {
		name           string
		transformation map[string]interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "valid transformations",
			transformation: map[string]interface{}{
				"name":  "$.name",
				"users": "$.users[*].name",
				"count": "$.total",
			},
			expectError: false,
		},
		{
			name:           "nil transformation",
			transformation: nil,
			expectError:    false,
		},
		{
			name:           "empty transformation",
			transformation: map[string]interface{}{},
			expectError:    false,
		},
		{
			name: "invalid JSONPath",
			transformation: map[string]interface{}{
				"invalid": "$.users[",
			},
			expectError: true,
			errorMsg:    "invalid JSONPath expression",
		},
		{
			name: "non-string path",
			transformation: map[string]interface{}{
				"invalid": 123,
			},
			expectError: true,
			errorMsg:    "transformation path for key 'invalid' must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer.ValidateTransformation(tt.transformation)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSONPathTransformer_ExtractValue(t *testing.T) {
	transformer := &JSONPathTransformer{}

	sampleData := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"items": []interface{}{1, 2, 3},
	}

	tests := []struct {
		name        string
		data        interface{}
		pathExpr    string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "extract nested field",
			data:     sampleData,
			pathExpr: "$.user.name",
			expected: "John",
		},
		{
			name:     "extract array element",
			data:     sampleData,
			pathExpr: "$.items[1]",
			expected: 2,
		},
		{
			name:     "extract root",
			data:     sampleData,
			pathExpr: "$",
			expected: sampleData,
		},
		{
			name:     "non-existent path",
			data:     sampleData,
			pathExpr: "$.nonexistent",
			expected: nil,
		},
		{
			name:        "invalid path",
			data:        sampleData,
			pathExpr:    "$.user[",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.ExtractValue(tt.data, tt.pathExpr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestJSONPathTransformer_TransformArray(t *testing.T) {
	transformer := &JSONPathTransformer{}

	arrayData := []interface{}{
		map[string]interface{}{
			"id":   1,
			"name": "John",
			"profile": map[string]interface{}{
				"age": 30,
			},
		},
		map[string]interface{}{
			"id":   2,
			"name": "Jane",
			"profile": map[string]interface{}{
				"age": 25,
			},
		},
	}

	tests := []struct {
		name           string
		data           interface{}
		transformation map[string]interface{}
		expected       []interface{}
		expectError    bool
		errorMsg       string
	}{
		{
			name: "transform array elements",
			data: arrayData,
			transformation: map[string]interface{}{
				"user_id":   "$.id",
				"user_name": "$.name",
				"user_age":  "$.profile.age",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_id":   1,
					"user_name": "John",
					"user_age":  30,
				},
				map[string]interface{}{
					"user_id":   2,
					"user_name": "Jane",
					"user_age":  25,
				},
			},
		},
		{
			name: "non-array data",
			data: map[string]interface{}{"key": "value"},
			transformation: map[string]interface{}{
				"key": "$.key",
			},
			expectError: true,
			errorMsg:    "data is not an array or slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformArray(tt.data, tt.transformation)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestJSONPathTransformer_ComplexTransformations(t *testing.T) {
	transformer := NewJSONPathTransformer()

	// Complex nested data structure
	complexData := map[string]interface{}{
		"response": map[string]interface{}{
			"data": map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"id": 1,
						"profile": map[string]interface{}{
							"personal": map[string]interface{}{
								"firstName": "John",
								"lastName":  "Doe",
							},
							"contact": map[string]interface{}{
								"emails": []interface{}{
									"john@work.com",
									"john@personal.com",
								},
							},
						},
					},
					map[string]interface{}{
						"id": 2,
						"profile": map[string]interface{}{
							"personal": map[string]interface{}{
								"firstName": "Jane",
								"lastName":  "Smith",
							},
							"contact": map[string]interface{}{
								"emails": []interface{}{
									"jane@work.com",
								},
							},
						},
					},
				},
			},
			"meta": map[string]interface{}{
				"total": 2,
				"page":  1,
			},
		},
	}

	transformation := map[string]interface{}{
		"total_users":    "$.response.meta.total",
		"first_names":    "$.response.data.users[*].profile.personal.firstName",
		"primary_emails": "$.response.data.users[*].profile.contact.emails[0]",
		"user_count":     "$.response.meta.total",
	}

	expected := map[string]interface{}{
		"total_users":    2,
		"first_names":    []interface{}{"John", "Jane"},
		"primary_emails": []interface{}{"john@work.com", "jane@work.com"}, // New JSONPath library correctly returns first email from each user
		"user_count":     2,
	}

	result, err := transformer.Transform(complexData, transformation)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
