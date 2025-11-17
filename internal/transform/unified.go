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
		jqTransformer: &JQTransformer{},
	}
}

// Transform applies jq transformation (legacy method for backward compatibility)
func (ut *UnifiedTransformer) Transform(data interface{}, jqQuery string) (interface{}, error) {
	return ut.jqTransformer.TransformWithQuery(data, jqQuery)
}

// TransformRequest applies transformation based on the proxy request configuration
func (ut *UnifiedTransformer) TransformRequest(data interface{}, req *models.ProxyRequest) (interface{}, error) {
	if req.TransformationMode != models.TransformationModeJQ {
		return nil, fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
	return ut.jqTransformer.TransformWithQuery(data, req.JQQuery)
}

// ValidateTransformation validates transformation configuration
func (ut *UnifiedTransformer) ValidateTransformation(req *models.ProxyRequest) error {
	if req.TransformationMode != models.TransformationModeJQ {
		return fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
	return ut.jqTransformer.ValidateQuery(req.JQQuery)
}

// GetJQTransformer returns the jq transformer
func (ut *UnifiedTransformer) GetJQTransformer() *JQTransformer {
	return ut.jqTransformer
}
