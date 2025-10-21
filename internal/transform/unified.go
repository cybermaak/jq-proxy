package transform

import (
	"fmt"

	"jq-proxy-service/internal/models"
)

// UnifiedTransformer handles both JSONPath and jq transformations
type UnifiedTransformer struct {
	jsonPathTransformer *JSONPathTransformer
	jqTransformer       *JQTransformer
}

// NewUnifiedTransformer creates a new unified transformer
func NewUnifiedTransformer() *UnifiedTransformer {
	return &UnifiedTransformer{
		jsonPathTransformer: &JSONPathTransformer{},
		jqTransformer:       &JQTransformer{},
	}
}

// Transform applies transformation based on the request configuration
func (ut *UnifiedTransformer) Transform(data interface{}, transformation map[string]interface{}) (interface{}, error) {
	// This is the legacy method for JSONPath transformations
	return ut.jsonPathTransformer.Transform(data, transformation)
}

// TransformRequest applies transformation based on the proxy request configuration
func (ut *UnifiedTransformer) TransformRequest(data interface{}, req *models.ProxyRequest) (interface{}, error) {
	switch req.TransformationMode {
	case models.TransformationModeJQ:
		return ut.jqTransformer.TransformWithQuery(data, req.JQQuery)
	case models.TransformationModeJSONPath, "":
		return ut.jsonPathTransformer.Transform(data, req.Transformation)
	default:
		return nil, fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
}

// ValidateTransformation validates transformation configuration
func (ut *UnifiedTransformer) ValidateTransformation(req *models.ProxyRequest) error {
	switch req.TransformationMode {
	case models.TransformationModeJQ:
		return ut.jqTransformer.ValidateQuery(req.JQQuery)
	case models.TransformationModeJSONPath, "":
		return ut.jsonPathTransformer.ValidateTransformation(req.Transformation)
	default:
		return fmt.Errorf("unsupported transformation mode: %s", req.TransformationMode)
	}
}

// GetJSONPathTransformer returns the JSONPath transformer for backward compatibility
func (ut *UnifiedTransformer) GetJSONPathTransformer() *JSONPathTransformer {
	return ut.jsonPathTransformer
}

// GetJQTransformer returns the jq transformer
func (ut *UnifiedTransformer) GetJQTransformer() *JQTransformer {
	return ut.jqTransformer
}
