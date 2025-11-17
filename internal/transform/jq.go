// Package transform provides response transformation capabilities using jq queries.
package transform

import (
	"fmt"

	"github.com/itchyny/gojq"
)

// JQTransformer implements jq-based transformations
type JQTransformer struct{}

// NewJQTransformer creates a new jq transformer
func NewJQTransformer() *JQTransformer {
	return &JQTransformer{}
}

// TransformWithQuery applies a jq query to the input data
func (jt *JQTransformer) TransformWithQuery(data any, query string) (any, error) {
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
	switch {
	case len(results) == 0:
		return nil, nil
	case len(results) == 1:
		return results[0], nil
	default:
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
