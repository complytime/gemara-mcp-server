package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	if layer < 1 || layer > 4 {
		return mcp.NewToolResultErrorf("layer must be between 1 and 4, got %d", layer), nil
	}

	// Perform CUE validation
	validationResult := g.PerformCUEValidation(yamlContent, layer)

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
		
		// Add common mistake suggestions based on error patterns
		suggestions := g.extractCommonMistakeSuggestions(validationResult, layer)
		if len(suggestions) > 0 {
			result += fmt.Sprintf("## Common Issues & Suggestions\n\n")
			for i, suggestion := range suggestions {
				result += fmt.Sprintf("%d. **%s**\n   %s\n\n", i+1, suggestion.Title, suggestion.Description)
			}
		}
		
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

	// Download all required schema files (common + layer-specific)
	commonSchemas := []string{"base.cue", "metadata.cue", "mapping.cue"}
	schemaFiles := make([]string, 0, len(commonSchemas)+1)

	// Download common schema files
	for _, schemaName := range commonSchemas {
		schemaContent, err := g.getCommonCUESchema(schemaName)
		if err != nil {
			result.Valid = false
			result.Error = fmt.Sprintf("Failed to load common schema %s: %v", schemaName, err)
			return result
		}
		schemaPath := filepath.Join(tmpDir, schemaName)
		if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
			result.Valid = false
			result.Error = fmt.Sprintf("Failed to write schema file %s: %v", schemaName, err)
			return result
		}
		schemaFiles = append(schemaFiles, schemaPath)
	}

	// Download layer-specific schema
	layerSchemaContent, err := g.getCUESchema(layer)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to load CUE schema for layer %d: %v", layer, err)
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

// CommonMistakeSuggestion represents a suggestion for fixing a common mistake
type CommonMistakeSuggestion struct {
	Title       string
	Description string
}

// extractCommonMistakeSuggestions analyzes validation errors and provides suggestions
func (g *GemaraAuthoringTools) extractCommonMistakeSuggestions(validationResult ValidationResult, layer int) []CommonMistakeSuggestion {
	suggestions := []CommonMistakeSuggestion{}
	
	// Combine all error messages for pattern matching
	allErrors := validationResult.Error
	for _, err := range validationResult.Errors {
		allErrors += " " + err
	}
	
	// Pattern matching for common mistakes
	errorLower := fmt.Sprintf("%s %s", validationResult.Error, allErrors)
	
	// Missing required fields
	if containsAny(errorLower, []string{"missing", "required", "not found"}) {
		if containsAny(errorLower, []string{"metadata", "id"}) {
			suggestions = append(suggestions, CommonMistakeSuggestion{
				Title:       "Missing metadata.id",
				Description: "Every Gemara artifact must have a unique `metadata.id` field. Use lowercase letters, numbers, hyphens, and underscores only (e.g., `my-guidance-v1`).",
			})
		}
		if containsAny(errorLower, []string{"metadata", "title"}) {
			suggestions = append(suggestions, CommonMistakeSuggestion{
				Title:       "Missing metadata.title",
				Description: "The `metadata.title` field is required. Provide a human-readable title for your artifact.",
			})
		}
		if layer == 1 && containsAny(errorLower, []string{"categories", "category"}) {
			suggestions = append(suggestions, CommonMistakeSuggestion{
				Title:       "Missing categories",
				Description: "Layer 1 Guidance documents must have at least one `category` with at least one `guideline`. Each category needs an `id` and `title`, and each guideline needs an `id` and `title`.",
			})
		}
		if layer == 2 && containsAny(errorLower, []string{"controls", "control"}) {
			suggestions = append(suggestions, CommonMistakeSuggestion{
				Title:       "Missing controls",
				Description: "Layer 2 Control Catalogs must have at least one `control` entry. Each control needs an `id`, `name`, and `technology` field.",
			})
		}
	}
	
	// Date format issues
	if containsAny(errorLower, []string{"date", "publication-date"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "Invalid date format",
			Description: "Dates must be in ISO 8601 format: `YYYY-MM-DD` (e.g., `2024-01-15`). Ensure the `publication-date` field uses this format.",
		})
	}
	
	// Document type issues
	if containsAny(errorLower, []string{"document-type", "document type"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "Invalid document-type",
			Description: "The `metadata.document-type` field must be one of: `Framework`, `Standard`, or `Guideline`. Check your spelling and capitalization.",
		})
	}
	
	// Type mismatches
	if containsAny(errorLower, []string{"cannot use", "incompatible", "type", "string", "number", "list"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "Type mismatch",
			Description: "Check that field types match the schema. Common issues: use strings (with quotes) for text fields, arrays (with `-`) for lists, and ensure nested objects are properly indented.",
		})
	}
	
	// YAML syntax issues
	if containsAny(errorLower, []string{"yaml", "parse", "syntax", "invalid"}) && !containsAny(errorLower, []string{"validation"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "YAML syntax error",
			Description: "Check your YAML syntax: ensure proper indentation (use spaces, not tabs), quote strings with special characters, and use `-` for array items. Validate your YAML with a YAML linter if needed.",
		})
	}
	
	// Reference issues
	if containsAny(errorLower, []string{"reference", "guideline-mapping", "layer1", "layer2", "control"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "Invalid reference",
			Description: "When referencing other artifacts, ensure the referenced IDs exist. Use `list_layer1_guidance()` or `list_layer2_controls()` to see available IDs. Then use `validate_artifact_references()` to verify all references are valid.",
		})
	}
	
	// Applicability issues
	if containsAny(errorLower, []string{"applicability", "jurisdiction", "technology-domain", "industry-sector"}) {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "Applicability field issues",
			Description: "The `metadata.applicability` object should contain arrays: `jurisdictions`, `technology-domains`, and `industry-sectors`. Each should be a list of strings (e.g., `- \"United States\"`).",
		})
	}
	
	// If no specific suggestions, provide general guidance
	if len(suggestions) == 0 {
		suggestions = append(suggestions, CommonMistakeSuggestion{
			Title:       "General validation error",
			Description: fmt.Sprintf("Review the error messages above carefully. Use `get_layer_schema_info(layer=%d)` to see the complete schema requirements. Check the `create-layer%d` prompt for examples.", layer, layer),
		})
	}
	
	return suggestions
}

// containsAny checks if a string contains any of the given substrings (case-insensitive)
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(substr) > 0 && len(s) >= len(substr) {
			// Simple substring check (case-insensitive)
			for i := 0; i <= len(s)-len(substr); i++ {
				if len(s) >= i+len(substr) {
					match := true
					for j := 0; j < len(substr); j++ {
						if toLower(s[i+j]) != toLower(substr[j]) {
							match = false
							break
						}
					}
					if match {
						return true
					}
				}
			}
		}
	}
	return false
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

// getCUESchema fetches or returns cached CUE schema for a layer
func (g *GemaraAuthoringTools) getCUESchema(layer int) (string, error) {
	// Check cache first
	if schema, ok := g.schemaCache[layer]; ok {
		return schema, nil
	}

	// Fetch schema from GitHub
	schemaURL := fmt.Sprintf("https://raw.githubusercontent.com/ossf/gemara/main/schemas/layer-%d.cue", layer)
	
	resp, err := http.Get(schemaURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch schema from %s: %w", schemaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch schema: HTTP %d", resp.StatusCode)
	}

	schemaBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read schema: %w", err)
	}

	schemaContent := string(schemaBytes)
	
	// Cache the schema
	g.schemaCache[layer] = schemaContent

	return schemaContent, nil
}

// getCommonCUESchema fetches or returns cached common CUE schema files
func (g *GemaraAuthoringTools) getCommonCUESchema(schemaName string) (string, error) {
	// Use a cache key that includes the schema name
	cacheKey := -1000 // Use negative numbers for common schemas to avoid conflicts with layer numbers
	if schemaName == "base.cue" {
		cacheKey = -1
	} else if schemaName == "metadata.cue" {
		cacheKey = -2
	} else if schemaName == "mapping.cue" {
		cacheKey = -3
	}

	// Check cache first
	if schema, ok := g.schemaCache[cacheKey]; ok {
		return schema, nil
	}

	// Fetch schema from GitHub
	schemaURL := fmt.Sprintf("https://raw.githubusercontent.com/ossf/gemara/main/schemas/%s", schemaName)
	
	resp, err := http.Get(schemaURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch schema from %s: %w", schemaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch schema: HTTP %d", resp.StatusCode)
	}

	schemaBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read schema: %w", err)
	}

	schemaContent := string(schemaBytes)
	
	// Cache the schema
	g.schemaCache[cacheKey] = schemaContent

	return schemaContent, nil
}

// handleGetLayerSchemaInfo provides information about a layer schema
func (g *GemaraAuthoringTools) handleGetLayerSchemaInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := request.GetInt("layer", 0)

	if layer < 1 || layer > 4 {
		return mcp.NewToolResultErrorf("layer must be between 1 and 4, got %d", layer), nil
	}

	var info string
	switch layer {
	case 1:
		info = `Layer 1: Guidance Schema Information

Purpose: High-level guidance on cybersecurity measures from industry groups, government agencies, or standards bodies.

Key Fields:
- name: Name of the guidance document (required)
- description: Description of the guidance (required)
- author: Author or organization (optional)
- version: Version of the guidance (optional)
- guidelines: Array of guideline objects (optional)

Schema Location: https://github.com/ossf/gemara/blob/main/schemas/layer-1.cue

Examples: NIST Cybersecurity Framework, ISO 27001, PCI DSS, HIPAA, GDPR`
	case 2:
		info = `Layer 2: Controls Schema Information

Purpose: Technology-specific, threat-informed security controls.

Key Fields:
- control_id: Unique identifier for the control (required)
- name: Name of the control (required)
- description: Description of what the control does (required)
- technology: Technology this control applies to (optional)
- threats: Array of threat identifiers this control mitigates (optional)
- layer1_references: Array of Layer 1 guidance IDs this control maps to (optional)

Schema Location: https://github.com/ossf/gemara/blob/main/schemas/layer-2.cue

Examples: CIS Benchmarks, FINOS Common Cloud Controls, OSPS Baseline`
	case 3:
		info = `Layer 3: Policy Schema Information

Purpose: Risk-informed governance rules tailored to an organization.

Key Fields:
- policy_id: Unique identifier for the policy (required)
- name: Name of the policy (required)
- description: Description of the policy (required)
- organization: Organization this policy applies to (optional)
- layer2_controls: Array of Layer 2 control IDs this policy references (optional)
- risk_appetite: Risk appetite level - low, medium, or high (optional)

Schema Location: https://github.com/ossf/gemara/blob/main/schemas/layer-3.cue

Note: Policies must consider organization-specific risk appetite and risk-acceptance.`
	case 4:
		info = `Layer 4: Evaluation Schema Information

Purpose: Inspection of code, configurations, and deployments.

Key Fields:
- evaluation_id: Unique identifier for the evaluation (required)
- name: Name of the evaluation (required)
- description: Description of what is being evaluated (optional)
- layer2_controls: Array of Layer 2 control IDs this evaluation assesses (optional)
- layer3_policies: Array of Layer 3 policy IDs this evaluation validates (optional)
- target_type: Type of target - code, configuration, or deployment (optional)

Schema Location: https://github.com/ossf/gemara/blob/main/schemas/layer-4.cue

Note: Evaluations may be built based on outputs from layers 2 or 3.`
	}

	return mcp.NewToolResultText(info), nil
}
