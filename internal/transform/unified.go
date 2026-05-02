// Package transform provides response transformation capabilities using jq queries.
package transform

import (
	"fmt"

	"jq-proxy-service/internal/models"
)

// UnifiedTransformer handles jq transformations
type UnifiedTransformer struct {
	jqTransformer *JQTransformer
}

// NewUnifiedTransformer creates a new unified transformer
func NewUnifiedTransformer() *UnifiedTransformer {
	return &UnifiedTransformer{
		jqTransformer: NewJQTransformer(),
	}
}

// TransformRequest applies transformation based on the proxy request configuration
func (ut *UnifiedTransformer) TransformRequest(data interface{}, req *models.ProxyRequest) (interface{}, error) {
	switch req.TransformationMode {
	case models.TransformationModeJQ:
		return ut.jqTransformer.TransformWithQuery(data, req.JQQuery)
	case models.TransformationModeNone:
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
}

// ValidateTransformation validates transformation configuration
func (ut *UnifiedTransformer) ValidateTransformation(req *models.ProxyRequest) error {
	switch req.TransformationMode {
	case models.TransformationModeJQ:
		return ut.jqTransformer.ValidateQuery(req.JQQuery)
	case models.TransformationModeNone:
		return nil
	default:
		return fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
}

// GetJQTransformer returns the jq transformer
func (ut *UnifiedTransformer) GetJQTransformer() *JQTransformer {
	return ut.jqTransformer
}
