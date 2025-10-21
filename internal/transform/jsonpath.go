package transform

import (
	"fmt"
	"reflect"
	"strings"

	"jq-proxy-service/internal/models"

	"github.com/theory/jsonpath"
)

// JSONPathTransformer implements ResponseTransformer using JSONPath expressions
type JSONPathTransformer struct{}

// NewJSONPathTransformer creates a new JSONPath transformer
func NewJSONPathTransformer() models.ResponseTransformer {
	return &JSONPathTransformer{}
}

// Transform applies JSONPath transformations to the input data
func (jt *JSONPathTransformer) Transform(data interface{}, transformation map[string]interface{}) (interface{}, error) {
	if transformation == nil {
		return data, nil
	}

	if len(transformation) == 0 {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})

	for key, pathExpr := range transformation {
		// Convert path expression to string
		pathStr, ok := pathExpr.(string)
		if !ok {
			return nil, fmt.Errorf("transformation path for key '%s' must be a string, got %T", key, pathExpr)
		}

		// Apply JSONPath expression
		value, err := jt.applyJSONPath(data, pathStr)
		if err != nil {
			return nil, fmt.Errorf("failed to apply JSONPath '%s' for key '%s': %w", pathStr, key, err)
		}

		result[key] = value
	}

	return result, nil
}

// applyJSONPath applies a single JSONPath expression to the data
func (jt *JSONPathTransformer) applyJSONPath(data interface{}, pathExpr string) (interface{}, error) {
	// Handle special case for root path
	if pathExpr == "$" {
		return data, nil
	}

	// Parse the JSONPath expression
	path, err := jsonpath.Parse(pathExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid JSONPath expression: %w", err)
	}

	// Execute the JSONPath query
	results := path.Select(data)

	// Handle different result scenarios
	if len(results) == 0 {
		return nil, nil // Path not found, return nil
	} else if len(results) == 1 {
		return results[0], nil // Single result
	} else {
		// Multiple results, return as array
		return []interface{}(results), nil
	}
}

// isNotFoundError checks if the error indicates a path was not found
func isNotFoundError(err error) bool {
	// The theory/jsonpath library returns specific error messages for not found cases
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "no such") ||
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "index out of range") ||
		strings.Contains(errStr, "key error")
}

// TransformWithFallback applies transformations with fallback values for missing paths
func (jt *JSONPathTransformer) TransformWithFallback(data interface{}, transformation map[string]interface{}, fallbacks map[string]interface{}) (interface{}, error) {
	if len(transformation) == 0 {
		return data, nil
	}

	result := make(map[string]interface{})

	for key, pathExpr := range transformation {
		pathStr, ok := pathExpr.(string)
		if !ok {
			return nil, fmt.Errorf("transformation path for key '%s' must be a string, got %T", key, pathExpr)
		}

		value, err := jt.applyJSONPath(data, pathStr)
		if err != nil {
			return nil, fmt.Errorf("failed to apply JSONPath '%s' for key '%s': %w", pathStr, key, err)
		}

		// Use fallback if value is nil and fallback exists
		if value == nil && fallbacks != nil {
			if fallbackValue, exists := fallbacks[key]; exists {
				value = fallbackValue
			}
		}

		result[key] = value
	}

	return result, nil
}

// ValidateTransformation validates that all JSONPath expressions in the transformation are valid
func (jt *JSONPathTransformer) ValidateTransformation(transformation map[string]interface{}) error {
	if transformation == nil {
		return nil
	}

	for key, pathExpr := range transformation {
		pathStr, ok := pathExpr.(string)
		if !ok {
			return fmt.Errorf("transformation path for key '%s' must be a string, got %T", key, pathExpr)
		}

		// Try to parse the JSONPath expression to validate syntax
		_, err := jsonpath.Parse(pathStr)
		if err != nil {
			return fmt.Errorf("invalid JSONPath expression for key '%s': %w", key, err)
		}
	}

	return nil
}

// ExtractValue extracts a single value using JSONPath
func (jt *JSONPathTransformer) ExtractValue(data interface{}, pathExpr string) (interface{}, error) {
	return jt.applyJSONPath(data, pathExpr)
}

// TransformArray applies transformation to each element in an array
func (jt *JSONPathTransformer) TransformArray(data interface{}, transformation map[string]interface{}) (interface{}, error) {
	// Check if data is an array/slice
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Slice && dataValue.Kind() != reflect.Array {
		return nil, fmt.Errorf("data is not an array or slice, got %T", data)
	}

	// Convert to interface slice
	dataSlice, ok := data.([]interface{})
	if !ok {
		// Try to convert
		result := make([]interface{}, dataValue.Len())
		for i := 0; i < dataValue.Len(); i++ {
			result[i] = dataValue.Index(i).Interface()
		}
		dataSlice = result
	}

	// Transform each element
	transformedArray := make([]interface{}, len(dataSlice))
	for i, item := range dataSlice {
		transformed, err := jt.Transform(item, transformation)
		if err != nil {
			return nil, fmt.Errorf("failed to transform array element at index %d: %w", i, err)
		}
		transformedArray[i] = transformed
	}

	return transformedArray, nil
}
