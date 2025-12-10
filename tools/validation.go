package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/mark3labs/mcp-go/mcp"
)

// ValidationResult holds the result of CUE validation
type ValidationResult struct {
	Valid  bool
	Error  string
	Errors []string
}

// handleValidateGemaraYAML validates YAML content against a layer schema using CUE
func (g *GemaraAuthoringTools) handleValidateGemaraYAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	layer := request.GetInt("layer", 0)
	outputFormat := request.GetString("output_format", "text")

	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	if layer < 1 || layer > 4 {
		return mcp.NewToolResultErrorf("layer must be between 1 and 4, got %d", layer), nil
	}

	// Perform CUE validation
	validationResult := g.PerformCUEValidation(yamlContent, layer)

	// Handle JSON format output
	if outputFormat == "json" {
		jsonResult := map[string]interface{}{
			"valid":  validationResult.Valid,
			"layer":  layer,
			"error":  validationResult.Error,
			"errors": validationResult.Errors,
			"schema": map[string]string{
				"url":        fmt.Sprintf("https://github.com/ossf/gemara/blob/main/schemas/layer-%d.cue", layer),
				"repository": "https://github.com/ossf/gemara/tree/main/schemas",
			},
		}
		jsonBytes, err := json.MarshalIndent(jsonResult, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		// Convert bytes to string - json.Marshal returns []byte
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// Default: text format
	// Build comprehensive validation result
	result := fmt.Sprintf(`# Gemara Layer %d Validation Report

## CUE Schema Validation
`, layer)

	if validationResult.Valid {
		result += "✅ CUE validation PASSED\n\n"
		result += fmt.Sprintf("The YAML content is valid according to the Layer %d CUE schema.\n\n", layer)
	} else {
		result += "❌ CUE validation FAILED\n\n"
		if validationResult.Error != "" {
			result += fmt.Sprintf("**Validation Error:**\n```\n%s\n```\n\n", validationResult.Error)
		}
		if len(validationResult.Errors) > 0 {
			result += "**Detailed Errors:**\n"
			for i, err := range validationResult.Errors {
				result += fmt.Sprintf("  %d. %s\n", i+1, err)
			}
			result += "\n"
		}
	}

	result += fmt.Sprintf("## Schema Information\n\n")
	result += fmt.Sprintf("- **Schema URL**: https://github.com/ossf/gemara/blob/main/schemas/layer-%d.cue\n", layer)
	result += fmt.Sprintf("- **Schema Repository**: https://github.com/ossf/gemara/tree/main/schemas\n\n")

	if !validationResult.Valid {
		result += fmt.Sprintf("## Your YAML Content\n\n")
		result += fmt.Sprintf("```yaml\n")
		result += fmt.Sprintf("%s\n", yamlContent)
		result += fmt.Sprintf("```\n\n")

		result += fmt.Sprintf("## Next Steps\n\n")
		result += fmt.Sprintf("1. Review the validation errors above\n")
		result += fmt.Sprintf("2. Check the suggestions section for common fixes\n")
		result += fmt.Sprintf("3. Ensure all required fields are present (use `get_layer_schema_info` with layer=%d)\n", layer)
		result += fmt.Sprintf("4. Verify field types match the schema requirements\n")
		result += fmt.Sprintf("5. Check that references use valid IDs (use `validate_artifact_references`)\n")
		result += fmt.Sprintf("6. Review examples in the `create-layer%d` prompt\n\n", layer)
	}

	return mcp.NewToolResultText(result), nil
}

// PerformCUEValidation performs CUE schema validation on YAML content
// This is exported so it can be used by validation scripts
func (g *GemaraAuthoringTools) PerformCUEValidation(yamlContent string, layer int) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Create a temporary directory for schema and data files
	tmpDir, err := os.MkdirTemp("", "gemara-validation-*")
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to create temporary directory: %v", err)
		return result
	}
	defer os.RemoveAll(tmpDir)

	// Load schemas using resource system (ensures consistency with MCP resources)
	// CUE requires file paths, so we write the schema content to temporary files
	commonSchemas := []struct {
		name     string
		resource string
	}{
		{"base.cue", "gemara://schema/common/base"},
		{"metadata.cue", "gemara://schema/common/metadata"},
		{"mapping.cue", "gemara://schema/common/mapping"},
	}
	schemaFiles := make([]string, 0, len(commonSchemas)+1)

	// Load common schema files using resource system
	for _, schema := range commonSchemas {
		schemaContent, err := g.getSchemaResourceContent(schema.resource)
		if err != nil {
			result.Valid = false
			result.Error = fmt.Sprintf("Failed to load schema resource %s: %v", schema.resource, err)
			return result
		}

		schemaPath := filepath.Join(tmpDir, schema.name)
		if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
			result.Valid = false
			result.Error = fmt.Sprintf("Failed to write schema file %s: %v", schema.name, err)
			return result
		}
		schemaFiles = append(schemaFiles, schemaPath)
	}

	// Load layer-specific schema using resource system
	if layer < 1 || layer > 4 {
		result.Valid = false
		result.Error = fmt.Sprintf("Invalid layer: %d (must be 1-4)", layer)
		return result
	}

	layerResourceURI := g.getLayerSchemaResourceURI(layer)
	layerSchemaContent, err := g.getSchemaResourceContent(layerResourceURI)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to load layer schema resource %s: %v", layerResourceURI, err)
		return result
	}

	layerSchemaPath := filepath.Join(tmpDir, fmt.Sprintf("layer-%d.cue", layer))
	if err := os.WriteFile(layerSchemaPath, []byte(layerSchemaContent), 0644); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to write layer schema file: %v", err)
		return result
	}
	schemaFiles = append(schemaFiles, layerSchemaPath)

	// Write YAML content to temporary file
	dataPath := filepath.Join(tmpDir, "data.yaml")
	if err := os.WriteFile(dataPath, []byte(yamlContent), 0644); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to write data file: %v", err)
		return result
	}

	// Load and validate using CUE
	ctx := cuecontext.New()

	// Load all schema files together
	schemaInstances := load.Instances(schemaFiles, &load.Config{
		Dir: tmpDir,
	})
	if len(schemaInstances) == 0 || schemaInstances[0].Err != nil {
		result.Valid = false
		if len(schemaInstances) > 0 && schemaInstances[0].Err != nil {
			result.Error = fmt.Sprintf("Failed to load schema: %v", schemaInstances[0].Err)
		} else {
			result.Error = "Failed to load schema: no instances returned"
		}
		return result
	}

	schemaValue := ctx.BuildInstance(schemaInstances[0])
	if err := schemaValue.Err(); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to build schema: %v", err)
		return result
	}

	// Load data
	dataInstances := load.Instances([]string{dataPath}, &load.Config{
		Dir: tmpDir,
	})
	if len(dataInstances) == 0 || dataInstances[0].Err != nil {
		result.Valid = false
		if len(dataInstances) > 0 && dataInstances[0].Err != nil {
			result.Error = fmt.Sprintf("Failed to load data: %v", dataInstances[0].Err)
		} else {
			result.Error = "Failed to load data: no instances returned"
		}
		return result
	}

	dataValue := ctx.BuildInstance(dataInstances[0])
	if err := dataValue.Err(); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to build data instance: %v", err)
		return result
	}

	// Unify schema and data
	unified := schemaValue.Unify(dataValue)
	if err := unified.Err(); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Schema unification failed: %v", err)
		return result
	}

	// Validate
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Validation failed: %v", err)

		// Extract detailed errors from the unified value
		// CUE errors are typically embedded in the value itself
		if unified.Err() != nil {
			result.Errors = append(result.Errors, unified.Err().Error())
		}

		// Also add the validation error
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	return result
}
