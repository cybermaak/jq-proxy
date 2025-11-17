package transform

import (
	"testing"

	"jq-proxy-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnifiedTransformer_TransformRequest_JQ(t *testing.T) {
	transformer := NewUnifiedTransformer()

	sampleData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": 1, "name": "John"},
			map[string]interface{}{"id": 2, "name": "Jane"},
		},
		"total": 2,
	}

	req := &models.ProxyRequest{
		Method:             "GET",
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{user_count: .total, user_names: [.users[].name]}",
	}

	expected := map[string]interface{}{
		"user_count": 2,
		"user_names": []interface{}{"John", "Jane"},
	}

	result, err := transformer.TransformRequest(sampleData, req)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestUnifiedTransformer_TransformRequest_DefaultMode(t *testing.T) {
	transformer := NewUnifiedTransformer()

	sampleData := map[string]interface{}{
		"name": "test",
	}

	req := &models.ProxyRequest{
		Method:             "GET",
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{result: .name}",
	}

	expected := map[string]interface{}{
		"result": "test",
	}

	result, err := transformer.TransformRequest(sampleData, req)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestUnifiedTransformer_TransformRequest_InvalidMode(t *testing.T) {
	transformer := NewUnifiedTransformer()

	req := &models.ProxyRequest{
		Method:             "GET",
		TransformationMode: "invalid",
	}

	result, err := transformer.TransformRequest(map[string]interface{}{}, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transformation mode")
	assert.Nil(t, result)
}

func TestUnifiedTransformer_ValidateTransformation_JQ(t *testing.T) {
	transformer := NewUnifiedTransformer()

	tests := []struct {
		name        string
		req         *models.ProxyRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid jq query",
			req: &models.ProxyRequest{
				TransformationMode: models.TransformationModeJQ,
				JQQuery:            ".data | map(.name)",
			},
			expectError: false,
		},
		{
			name: "invalid jq query",
			req: &models.ProxyRequest{
				TransformationMode: models.TransformationModeJQ,
				JQQuery:            ".data | map(",
			},
			expectError: true,
			errorMsg:    "invalid jq query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer.ValidateTransformation(tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnifiedTransformer_ValidateTransformation_InvalidMode(t *testing.T) {
	transformer := NewUnifiedTransformer()

	req := &models.ProxyRequest{
		TransformationMode: "invalid",
	}

	err := transformer.ValidateTransformation(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported transformation mode")
}

func TestUnifiedTransformer_GetTransformers(t *testing.T) {
	transformer := NewUnifiedTransformer()

	jqTransformer := transformer.GetJQTransformer()
	assert.NotNil(t, jqTransformer)
	assert.IsType(t, &JQTransformer{}, jqTransformer)
}
