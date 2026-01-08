package info

import (
	"context"
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/yaml"
	"github.com/complytime/gemara-mcp-server/internal/consts"
	"github.com/mark3labs/mcp-go/mcp"
)

// handleValidateGemaraYAML validates YAML content against a layer schema using CUE
func (g *GemaraInfoTools) handleValidateGemaraYAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	layer := request.GetInt("layer", 0)
	outputFormat := request.GetString("output_format", "text")

	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	if layer < consts.MinLayer || layer > consts.MaxLayer {
		return mcp.NewToolResultErrorf("layer must be between %d and %d, got %d", consts.MinLayer, consts.MaxLayer, layer), nil
	}

	// Perform CUE validation
	validationResult := g.PerformCUEValidation(yamlContent, layer)
	report := ValidationReport{
		ValidationResult: validationResult,
		Layer:            layer,
		SchemaVersion:    g.schemaVersion,
		Schema: struct {
			URL        string `json:"url"`
			Repository string `json:"repository"`
		}{
			URL:        fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-%d.cue", g.schemaVersion, layer),
			Repository: fmt.Sprintf("https://github.com/ossf/gemara/tree/%s/schemas", g.schemaVersion),
		},
	}

	// Handle JSON format output
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		// Convert bytes to string - json.Marshal returns []byte
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	return mcp.NewToolResultText(report.ToText(yamlContent)), nil
}

// PerformCUEValidation performs CUE schema validation on YAML content
// This is exported so it can be used by validation scripts
func (g *GemaraInfoTools) PerformCUEValidation(yamlContent string, layer int) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	const (
		base     = "gemara://schema/common/base"
		metadata = "gemara://schema/common/metadata"
		mapping  = "gemara://schema/common/mapping"
	)

	baseSchemaContent, err := g.getSchemaResourceContent(base)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("failed to load schema resource %s: %v", base, err)
		return result
	}

	metadataSchemaContent, err := g.getSchemaResourceContent(metadata)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("failed to load schema resource %s: %v", metadata, err)
		return result
	}

	mappingSchemaContent, err := g.getSchemaResourceContent(mapping)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to load schema resource %s: %v", mapping, err)
		return result
	}

	// Load layer-specific schema using resource system
	if layer < consts.MinLayer || layer > consts.MaxLayer {
		result.Valid = false
		result.Error = fmt.Sprintf("Invalid layer: %d (must be %d-%d)", layer, consts.MinLayer, consts.MaxLayer)
		return result
	}

	layerResourceURI := g.getLayerSchemaResourceURI(layer)
	layerSchemaContent, err := g.getSchemaResourceContent(layerResourceURI)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to load layer schema resource %s: %v", layerResourceURI, err)
		return result
	}

	// Create an Overlay
	// This maps "fake" filenames to the content strings.
	overlay := map[string]load.Source{
		"/base.cue":     load.FromBytes([]byte(baseSchemaContent)),
		"/metadata.cue": load.FromBytes([]byte(metadataSchemaContent)),
		"/mapping.cue":  load.FromBytes([]byte(mappingSchemaContent)),
		"/layer.cue":    load.FromBytes([]byte(layerSchemaContent)),
	}

	// 3. Configure the Loader
	cfg := &load.Config{
		Overlay: overlay,
		Dir:     "/", // The root of our fake filesystem
	}

	// 4. Load the instances
	// "." tells CUE to load the package found in the Dir ("/")
	buildInstances := load.Instances([]string{"."}, cfg)

	// Check for build/syntax errors in the schema itself
	if len(buildInstances) != 1 {
		result.Valid = false
		result.Error = fmt.Sprintf("expected 1 CUE package, found %d. Ensure all schema files define the same package name.", len(buildInstances))
		return result
	}
	if err := buildInstances[0].Err; err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("schema build failed: %v", err)
		return result
	}

	// 5. Build the Schema Instance
	ctx := cuecontext.New()
	schema := ctx.BuildInstance(buildInstances[0])
	if err := schema.Err(); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("schema compilation failed: %v", err)
		return result
	}

	// 6. Narrow down the schema based on the Layer
	var entryPoint cue.Value
	switch layer {
	case consts.Layer1:
		entryPoint = schema.LookupPath(cue.ParsePath("#GuidanceDocument"))
	case consts.Layer2:
		entryPoint = schema.LookupPath(cue.ParsePath("#Catalog"))
	case consts.Layer3:
		entryPoint = schema.LookupPath(cue.ParsePath("#Policy"))
	case consts.Layer4:
		entryPoint = schema.LookupPath(cue.ParsePath("#EvaluationLog"))
	}

	// If the lookup fails, we default back to the whole schema or error out
	if !entryPoint.Exists() {
		result.Valid = false
		result.Error = fmt.Sprintf("could not find entry point definition for layer %d", layer)
		return result
	}

	yamlFile, err := yaml.Extract("data.yml", yamlContent)
	if err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Failed to parse YAML: %v", err)
		return result
	}

	// Build the YAML as a CUE value
	data := ctx.BuildFile(yamlFile)
	if err := data.Err(); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("invalid data structure: %v", err)
		return result
	}

	// 8. Unify Schema with Data
	unified := entryPoint.Unify(data)

	// 9. Validate
	// Validate with Concrete(true) ensures all fields are filled
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		result.Valid = false
		result.Error = fmt.Sprintf("Validation failed: %v", err)
		result.Errors = append(result.Errors, result.Errors...)
		return result
	}
	return result
}

// ValidationResult holds the result of CUE validation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Error  string   `json:"error"`
	Errors []string `json:"errors"`
}

type ValidationReport struct {
	ValidationResult `json:,inline`
	Layer            int    `json:"layer"`
	SchemaVersion    string `json:"schema_version"`
	Schema           struct {
		URL        string `json:"url"`
		Repository string `json:"repository"`
	}
}

func (v ValidationReport) ToText(yamlContent string) string {
	result := fmt.Sprintf(`# Gemara Layer %d Validation Report

## CUE Schema Validation
`, v.Layer)

	if v.ValidationResult.Valid {
		result += "✅ CUE validation PASSED\n\n"
		result += fmt.Sprintf("The YAML content is valid according to the Layer %d CUE schema.\n\n", v.Layer)
	} else {
		result += "❌ CUE validation FAILED\n\n"
		if v.ValidationResult.Error != "" {
			result += fmt.Sprintf("**Validation Error:**\n```\n%s\n```\n\n", v.ValidationResult.Error)
		}
		if len(v.ValidationResult.Errors) > 0 {
			result += "**Detailed Errors:**\n"
			for i, err := range v.ValidationResult.Errors {
				result += fmt.Sprintf("  %d. %s\n", i+1, err)
			}
			result += "\n"
		}
	}

	result += fmt.Sprintf("## Schema Information\n\n")
	result += fmt.Sprintf("- **Schema Version**: %s\n", v.SchemaVersion)
	result += fmt.Sprintf("- **Schema URL**: https://github.com/ossf/gemara/blob/%s/schemas/layer-%d.cue\n", v.SchemaVersion, v.Layer)
	result += fmt.Sprintf("- **Schema Repository**: https://github.com/ossf/gemara/tree/%s/schemas\n\n", v.SchemaVersion)

	if !v.ValidationResult.Valid {
		result += fmt.Sprintf("## Your YAML Content\n\n")
		result += fmt.Sprintf("```yaml\n")
		result += fmt.Sprintf("%s\n", yamlContent)
		result += fmt.Sprintf("```\n\n")

		result += fmt.Sprintf("## Next Steps\n\n")
		result += fmt.Sprintf("1. Review the validation errors above\n")
		result += fmt.Sprintf("2. Check the suggestions section for common fixes\n")
		result += fmt.Sprintf("3. Ensure all required fields are present (use `get_layer_schema_info` with layer=%d)\n", v.Layer)
		result += fmt.Sprintf("4. Verify field types match the schema requirements\n")
		result += fmt.Sprintf("5. Check that references use valid IDs (use `validate_artifact_references`)\n")
		result += fmt.Sprintf("6. Review examples in the `create-layer%d` prompt\n\n", v.Layer)
	}
	return result
}
