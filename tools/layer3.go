package tools

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer3"
)

// REMOVED: handleCreateLayer3Policy - Use store_layer3_yaml instead
// REMOVED: handleCreatePolicyThroughScoping - Use store_layer3_yaml instead

// handleLoadLayer3FromFile loads a Layer 3 Policy document from a file using Gemara library

// handleLoadLayer3FromFile loads a Layer 3 Policy document from a file using Gemara library
func (g *GemaraAuthoringTools) handleLoadLayer3FromFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	policy := &layer3.PolicyDocument{}
	if err := policy.LoadFile(filePath); err != nil {
		return mcp.NewToolResultErrorf("Failed to load Layer 3 Policy from file: %v", err), nil
	}

	// Get policy ID from metadata
	policyID := ""
	if policy.Metadata.Id != "" {
		policyID = policy.Metadata.Id
	} else {
		return mcp.NewToolResultError("Loaded policy document missing metadata.id"), nil
	}

	// Store the loaded policy in Gemara types storage
	g.layer3Policies[policyID] = policy

	result := fmt.Sprintf("Successfully loaded Layer 3 Policy:\n")
	result += fmt.Sprintf("- Policy ID: %s\n", policyID)
	if policy.Metadata.Title != "" {
		result += fmt.Sprintf("- Title: %s\n", policy.Metadata.Title)
	}
	if policy.Metadata.Objective != "" {
		result += fmt.Sprintf("- Objective: %s\n", policy.Metadata.Objective)
	}
	result += fmt.Sprintf("\nPolicy loaded and available for querying.\n")

	return mcp.NewToolResultText(result), nil
}

// handleStoreLayer3YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 3 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer3YAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Validate with CUE first
	validationResult := g.PerformCUEValidation(yamlContent, 3)
	if !validationResult.Valid {
		errorMsg := "CUE validation failed:\n"
		if validationResult.Error != "" {
			errorMsg += validationResult.Error + "\n"
		}
		if len(validationResult.Errors) > 0 {
			errorMsg += "Errors:\n"
			for _, err := range validationResult.Errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
		}
		return mcp.NewToolResultError(errorMsg), nil
	}

	// Extract ID from YAML for storage
	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse YAML to extract metadata: %v", err), nil
	}

	var artifactID string
	if meta, ok := metadata["metadata"].(map[string]interface{}); ok {
		if id, ok := meta["id"].(string); ok {
			artifactID = id
		}
	}

	if artifactID == "" {
		return mcp.NewToolResultError("YAML content must include metadata.id"), nil
	}

	// Store raw YAML to disk
	if g.storage == nil {
		return mcp.NewToolResultError("Storage not available"), nil
	}

	storedID, err := g.storage.StoreRawYAML(3, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	result := fmt.Sprintf("Successfully stored and validated Layer 3 Policy:\n")
	result += fmt.Sprintf("- Policy ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nPolicy stored and available for querying.\n")

	return mcp.NewToolResultText(result), nil
}
