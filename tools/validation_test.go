// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validYAMLL1 = `metadata:
  id: test-guidance-1
  description: "A test guidance document for validation"
  author:
    id: test
    name: TEST
    type: Human
  version: "1.0"
document-type: "Standard"
title: "Test Guidance Document"
categories:
  - id: test-category
    title: "Test Category"
    description: "A test category"
    guidelines:
      - id: test-guideline
        title: "Test Guideline"
        objective: "Test objective"
        recommendations:
          - "Test recommendation"`

	validYAMLL2 = `metadata:
  id: test-catalog-1
  title: "Test Control Catalog"
  description: "A test control catalog"
  version: "1.0"

controls:
  - id: test-control-1
    title: "Test Control"
    objective: "Test control objective"
    requirements:
      - id: req-1
        description: "Test requirement"`

	validYAMLL3 = `metadata:
  id: test-policy-1
  description: "A test policy document"
  author:
    id: test
    name: TEST
    type: Human
  version: "1.0"
title: "Test Policy Document"
purpose: "Test Purpose"
"organization-id": "Test Organization"`
)

// TestPerformCUEValidation_ValidLayer1 tests validation with valid Layer 1 YAML
// Note: This test requires network access to fetch schemas from GitHub.
func TestPerformCUEValidation(t *testing.T) {
	g, err := NewGemaraAuthoringTools()
	require.NoError(t, err)
	tests := []struct {
		name        string
		layer       int
		yamlContent string
		wantValid   bool
	}{
		{
			name:        "Valid/Layer1",
			layer:       1,
			yamlContent: validYAMLL1,
			wantValid:   true,
		},
		{
			name:        "Invalid/Layer2",
			layer:       2,
			yamlContent: validYAMLL2,
			wantValid:   false,
		},
		{
			name:        "Valid/Layer3",
			layer:       3,
			yamlContent: validYAMLL3,
			wantValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.PerformCUEValidation(tt.yamlContent, tt.layer)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}
