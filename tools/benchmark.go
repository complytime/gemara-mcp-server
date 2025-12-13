package tools

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1/pipeline/parser"
	"github.com/ossf/gemara/layer1/pipeline/types"
)

// handleBenchmarkParser benchmarks parser performance
func (g *GemaraAuthoringTools) handleBenchmarkParser(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	iterations := 3
	if iter, ok := argsMap["iterations"].(float64); ok {
		iterations = int(iter)
		if iterations < 1 {
			iterations = 1
		}
		if iterations > 10 {
			iterations = 10 // Cap at 10 iterations
		}
	}

	parserType := "simple"
	if pt, ok := argsMap["parser_type"].(string); ok && pt != "" {
		parserType = pt
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

	// Configure parser
	config := types.ParserConfig{
		Provider:      parserType,
		TempDir:       tempDir,
		KeepTempFiles: false, // Always clean up temp files for benchmarking
	}

	// Run benchmarks
	var durations []time.Duration
	var errors []string
	var totalBlocks []int
	var totalPages []int

	for i := 0; i < iterations; i++ {
		// Create parser
		p, err := parser.NewParser(config)
		if err != nil {
			errors = append(errors, fmt.Sprintf("iteration %d: failed to create parser: %v", i+1, err))
			continue
		}

		// Parse PDF
		startTime := time.Now()
		doc, err := p.Parse(filePath)
		elapsed := time.Since(startTime)

		if err != nil {
			errors = append(errors, fmt.Sprintf("iteration %d: parsing failed: %v", i+1, err))
			continue
		}

		durations = append(durations, elapsed)
		totalBlocks = append(totalBlocks, countBlocks(doc))
		totalPages = append(totalPages, len(doc.Pages))
	}

	// Calculate statistics
	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration
	if len(durations) > 0 {
		minDuration = durations[0]
		maxDuration = durations[0]
		for _, d := range durations {
			totalDuration += d
			if d < minDuration {
				minDuration = d
			}
			if d > maxDuration {
				maxDuration = d
			}
		}
	}

	avgDuration := time.Duration(0)
	if len(durations) > 0 {
		avgDuration = totalDuration / time.Duration(len(durations))
	}

	// Prepare result
	result := map[string]interface{}{
		"file_path":       filePath,
		"parser_type":     parserType,
		"iterations":      iterations,
		"successful_runs": len(durations),
		"failed_runs":     len(errors),
		"errors":          errors,
		"statistics": map[string]interface{}{
			"min_duration_ms":   minDuration.Milliseconds(),
			"max_duration_ms":   maxDuration.Milliseconds(),
			"avg_duration_ms":   avgDuration.Milliseconds(),
			"total_duration_ms": totalDuration.Milliseconds(),
			"min_duration":      minDuration.String(),
			"max_duration":      maxDuration.String(),
			"avg_duration":      avgDuration.String(),
			"total_duration":    totalDuration.String(),
		},
	}

	if len(totalPages) > 0 {
		result["pages"] = totalPages[0] // Should be consistent across runs
	}
	if len(totalBlocks) > 0 {
		result["total_blocks"] = totalBlocks[0] // Should be consistent across runs
	}

	// Include individual run details if requested
	if includeDetails, ok := argsMap["include_details"].(bool); ok && includeDetails {
		details := make([]map[string]interface{}, len(durations))
		for i, d := range durations {
			details[i] = map[string]interface{}{
				"iteration":    i + 1,
				"duration_ms":  d.Milliseconds(),
				"duration":     d.String(),
				"pages":        totalPages[i],
				"total_blocks": totalBlocks[i],
			}
		}
		result["runs"] = details
	}

	// Marshal result
	var output string
	var err error
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
