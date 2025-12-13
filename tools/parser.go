package tools

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1/pipeline/converter"
	"github.com/ossf/gemara/layer1/pipeline/parser"
	"github.com/ossf/gemara/layer1/pipeline/segmenter"
	"github.com/ossf/gemara/layer1/pipeline/types"
	"github.com/ossf/gemara/layer1/pipeline/validator"
)

// handleSimpleParse parses a PDF file using the simple parser
func (g *GemaraAuthoringTools) handleSimpleParse(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsRaw := request.GetRawArguments()
	argsMap, ok := argsRaw.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}

	// Extract file path
	filePath, ok := argsMap["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return mcp.NewToolResultErrorf("file not found: %s", filePath), nil
	}

	// Extract optional parameters
	tempDir := ""
	if td, ok := argsMap["temp_dir"].(string); ok && td != "" {
		tempDir = td
	} else {
		// Use system temp directory
		tempDir = os.TempDir()
	}

	keepTempFiles := false
	if ktf, ok := argsMap["keep_temp_files"].(bool); ok {
		keepTempFiles = ktf
	}

	outputFormat := "json"
	if of, ok := argsMap["output_format"].(string); ok && of != "" {
		outputFormat = of
	}

	// Configure parser
	config := types.ParserConfig{
		Provider:      "simple",
		TempDir:       tempDir,
		KeepTempFiles: keepTempFiles,
	}

	// Create parser
	p, err := parser.NewParser(config)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to create parser: %v", err), nil
	}

	// Parse PDF
	startTime := time.Now()
	doc, err := p.Parse(filePath)
	elapsed := time.Since(startTime)

	if err != nil {
		return mcp.NewToolResultErrorf("parsing failed: %v", err), nil
	}

	// Prepare result
	result := map[string]interface{}{
		"success":      true,
		"file_path":    filePath,
		"parser":       doc.Metadata.Parser,
		"parsed_at":    doc.Metadata.ParsedAt.Format(time.RFC3339),
		"pages":        len(doc.Pages),
		"total_blocks": countBlocks(doc),
		"duration_ms":  elapsed.Milliseconds(),
		"duration":     elapsed.String(),
	}

	// Include full document if requested
	if includeFull, ok := argsMap["include_full_document"].(bool); ok && includeFull {
		result["document"] = doc
	} else {
		// Include summary statistics
		result["summary"] = map[string]interface{}{
			"pages":        len(doc.Pages),
			"total_blocks": countBlocks(doc),
			"block_types":  getBlockTypeCounts(doc),
		}
	}

	// Optionally validate if layer1_yaml_content is provided
	if layer1YAML, ok := argsMap["layer1_yaml_content"].(string); ok && layer1YAML != "" {
		validationResult := g.PerformCUEValidation(layer1YAML, 1)
		result["validation"] = map[string]interface{}{
			"valid":  validationResult.Valid,
			"errors": validationResult.Errors,
		}
		if validationResult.Error != "" {
			result["validation"].(map[string]interface{})["error"] = validationResult.Error
		}
	}

	// Marshal result
	var output string
	if outputFormat == "yaml" {
		output, err = marshalOutput(result, "yaml")
	} else {
		output, err = marshalOutput(result, "json")
	}
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal result: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

// handleParseAndValidate performs full pipeline: parse -> segment -> convert -> validate
func (g *GemaraAuthoringTools) handleParseAndValidate(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsRaw := request.GetRawArguments()
	argsMap, ok := argsRaw.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}

	// Extract file path
	filePath, ok := argsMap["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return mcp.NewToolResultErrorf("file not found: %s", filePath), nil
	}

	// Extract optional parameters
	parserType := "simple"
	if pt, ok := argsMap["parser_type"].(string); ok && pt != "" {
		parserType = pt
	}

	segmenterType := "generic"
	if st, ok := argsMap["segmenter_type"].(string); ok && st != "" {
		segmenterType = st
	}

	strict := false
	if s, ok := argsMap["strict"].(bool); ok {
		strict = s
	}

	tempDir := ""
	if td, ok := argsMap["temp_dir"].(string); ok && td != "" {
		tempDir = td
	} else {
		tempDir = os.TempDir()
	}

	outputFormat := "json"
	if of, ok := argsMap["output_format"].(string); ok && of != "" {
		outputFormat = of
	}

	// Step 1: Parse
	parseConfig := types.ParserConfig{
		Provider:      parserType,
		TempDir:       tempDir,
		KeepTempFiles: false,
	}

	parser, err := parser.NewParser(parseConfig)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to create parser: %v", err), nil
	}

	parseStart := time.Now()
	parsedDoc, err := parser.Parse(filePath)
	parseDuration := time.Since(parseStart)
	if err != nil {
		return mcp.NewToolResultErrorf("parsing failed: %v", err), nil
	}

	// Step 2: Segment
	segmentConfig := types.SegmenterConfig{
		DocumentType: segmenterType,
	}

	segmenter, err := segmenter.NewSegmenter(segmentConfig)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to create segmenter: %v", err), nil
	}

	segmentStart := time.Now()
	segmentedDoc, err := segmenter.Segment(parsedDoc)
	segmentDuration := time.Since(segmentStart)
	if err != nil {
		return mcp.NewToolResultErrorf("segmentation failed: %v", err), nil
	}

	// Step 3: Convert
	conv := converter.NewConverter()
	convertStart := time.Now()
	layer1Doc, err := conv.Convert(segmentedDoc)
	convertDuration := time.Since(convertStart)
	if err != nil {
		return mcp.NewToolResultErrorf("conversion failed: %v", err), nil
	}

	// Step 4: Validate
	validateStart := time.Now()
	var v *validator.Validator
	if strict {
		v = validator.NewValidator(validator.WithStrictMode(true))
	} else {
		v = validator.NewValidator()
	}
	validationResult := v.Validate(layer1Doc)
	validateDuration := time.Since(validateStart)

	// Also perform CUE validation
	yamlBytes, err := yaml.Marshal(layer1Doc)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal for CUE validation: %v", err), nil
	}
	cueValidationResult := g.PerformCUEValidation(string(yamlBytes), 1)

	// Build result
	result := map[string]interface{}{
		"file_path": filePath,
		"pipeline": map[string]interface{}{
			"parse": map[string]interface{}{
				"success":     true,
				"duration_ms": parseDuration.Milliseconds(),
				"duration":    parseDuration.String(),
				"pages":       len(parsedDoc.Pages),
				"blocks":      countBlocks(parsedDoc),
			},
			"segment": map[string]interface{}{
				"success":     true,
				"duration_ms": segmentDuration.Milliseconds(),
				"duration":    segmentDuration.String(),
				"categories":  len(segmentedDoc.Categories),
			},
			"convert": map[string]interface{}{
				"success":     true,
				"duration_ms": convertDuration.Milliseconds(),
				"duration":    convertDuration.String(),
			},
			"validate": map[string]interface{}{
				"success":     true,
				"duration_ms": validateDuration.Milliseconds(),
				"duration":    validateDuration.String(),
			},
		},
		"validation": map[string]interface{}{
			"programmatic": map[string]interface{}{
				"valid":  validationResult.Valid,
				"errors": len(validationResult.Errors),
			},
			"cue": map[string]interface{}{
				"valid":  cueValidationResult.Valid,
				"errors": len(cueValidationResult.Errors),
			},
		},
	}

	// Add error details if validation failed
	if !validationResult.Valid {
		errors := make([]map[string]interface{}, len(validationResult.Errors))
		for i, e := range validationResult.Errors {
			errors[i] = map[string]interface{}{
				"path":    e.Path,
				"message": e.Message,
				"value":   e.Value,
			}
		}
		result["validation"].(map[string]interface{})["programmatic"].(map[string]interface{})["error_details"] = errors
	}

	if !cueValidationResult.Valid {
		result["validation"].(map[string]interface{})["cue"].(map[string]interface{})["error_details"] = cueValidationResult.Errors
		if cueValidationResult.Error != "" {
			result["validation"].(map[string]interface{})["cue"].(map[string]interface{})["error"] = cueValidationResult.Error
		}
	}

	// Include Layer 1 document if requested
	if includeDoc, ok := argsMap["include_layer1_document"].(bool); ok && includeDoc {
		result["layer1_document"] = layer1Doc
	}

	// Marshal result
	var output string
	if outputFormat == "yaml" {
		output, err = marshalOutput(result, "yaml")
	} else {
		output, err = marshalOutput(result, "json")
	}
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal result: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

// countBlocks counts the total number of blocks in a parsed document
func countBlocks(doc *types.ParsedDocument) int {
	count := 0
	for _, page := range doc.Pages {
		count += len(page.Blocks)
	}
	return count
}

// getBlockTypeCounts returns a map of block type counts
func getBlockTypeCounts(doc *types.ParsedDocument) map[string]int {
	counts := make(map[string]int)
	for _, page := range doc.Pages {
		for _, block := range page.Blocks {
			counts[string(block.Type)]++
		}
	}
	return counts
}

// checkPdfToolsAvailable checks if required PDF tools are available
func checkPdfToolsAvailable() (map[string]bool, []string) {
	tools := map[string]string{
		"pdftotext": "pdftotext",
		"pdfinfo":   "pdfinfo",
	}

	available := make(map[string]bool)
	var missing []string

	for name, cmd := range tools {
		if _, err := exec.LookPath(cmd); err == nil {
			available[name] = true
		} else {
			available[name] = false
			missing = append(missing, name)
		}
	}

	return available, missing
}
