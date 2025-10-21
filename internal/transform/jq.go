package transform

import (
	"fmt"

	"jq-proxy-service/internal/models"

	"github.com/itchyny/gojq"
)

// JQTransformer implements ResponseTransformer using jq syntax
type JQTransformer struct{}

// NewJQTransformer creates a new jq transformer
func NewJQTransformer() models.ResponseTransformer {
	return &JQTransformer{}
}

// Transform applies jq transformation to the input data
func (jt *JQTransformer) Transform(data interface{}, transformation map[string]interface{}) (interface{}, error) {
	// For JQ transformer, we expect a single query string
	// This method is kept for interface compatibility but should use TransformWithQuery instead
	return nil, fmt.Errorf("JQ transformer requires using TransformWithQuery method")
}

// TransformWithQuery applies a jq query to the input data
func (jt *JQTransformer) TransformWithQuery(data interface{}, query string) (interface{}, error) {
	if query == "" {
		return data, nil
	}

	// Parse the jq query
	q, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query: %w", err)
	}

	// Compile the query for better performance
	code, err := gojq.Compile(q)
	if err != nil {
		return nil, fmt.Errorf("failed to compile jq query: %w", err)
	}

	// Execute the query
	iter := code.Run(data)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq query execution failed: %w", err)
		}
		results = append(results, v)
	}

	// Return single result if only one, otherwise return array
	if len(results) == 0 {
		return nil, nil
	} else if len(results) == 1 {
		return results[0], nil
	} else {
		return results, nil
	}
}

// ValidateQuery validates that a jq query is syntactically correct
func (jt *JQTransformer) ValidateQuery(query string) error {
	if query == "" {
		return nil
	}

	// Try to parse the jq query
	_, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("invalid jq query: %w", err)
	}

	return nil
}

// JQTransformerInterface extends ResponseTransformer with jq-specific methods
type JQTransformerInterface interface {
	models.ResponseTransformer
	TransformWithQuery(data interface{}, query string) (interface{}, error)
	ValidateQuery(query string) error
}
