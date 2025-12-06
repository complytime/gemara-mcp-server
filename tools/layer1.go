package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
)


// handleListLayer1Guidance lists all available Layer 1 Guidance documents
func (g *GemaraAuthoringTools) handleListLayer1Guidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get list from storage if available, otherwise use in-memory cache
	var entries []*ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(1)
	} else {
		// Fallback to in-memory cache
		for guidanceID, guidance := range g.layer1Guidance {
			entries = append(entries, &ArtifactIndexEntry{
				ID:    guidanceID,
				Layer: 1,
				Title: guidance.Metadata.Title,
			})
		}
	}

	totalCount := len(entries)
	if totalCount == 0 {
		return mcp.NewToolResultText("No Layer 1 Guidance documents available.\n\nUse store_layer1_yaml to store guidance documents or load_layer1_from_file to load from disk."), nil
	}

	result := fmt.Sprintf("# Available Layer 1 Guidance Documents\n\n")
	result += fmt.Sprintf("Total: %d guidance document(s)\n\n", totalCount)

	for _, entry := range entries {
		// Try to get full details from cache or storage
		var guidance *layer1.GuidanceDocument
		if gd, exists := g.layer1Guidance[entry.ID]; exists {
			guidance = gd
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
				if gd, ok := retrieved.(*layer1.GuidanceDocument); ok {
					guidance = gd
					// Update cache
					g.layer1Guidance[entry.ID] = guidance
				}
			}
		}

		result += fmt.Sprintf("## %s\n", entry.Title)
		result += fmt.Sprintf("- **ID**: `%s`\n", entry.ID)
		if guidance != nil {
			if guidance.Metadata.Description != "" {
				result += fmt.Sprintf("- **Description**: %s\n", guidance.Metadata.Description)
			}
			if guidance.Metadata.Author != "" {
				result += fmt.Sprintf("- **Author**: %s\n", guidance.Metadata.Author)
			}
			if guidance.Metadata.Version != "" {
				result += fmt.Sprintf("- **Version**: %s\n", guidance.Metadata.Version)
			}
		}
		result += "\n"
	}

	result += "\nUse `get_layer1_guidance` with a guidance_id to get full details.\n"
	result += "Use these guidance IDs in `guideline_mappings` when creating Layer 2 controls.\n"

	return mcp.NewToolResultText(result), nil
}

// handleGetLayer1Guidance gets detailed information about a specific Layer 1 Guidance
func (g *GemaraAuthoringTools) handleGetLayer1Guidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	guidanceID := request.GetString("guidance_id", "")
	_ = request.GetString("output_format", "yaml") // Output format handled via JSON marshaling

	if guidanceID == "" {
		return mcp.NewToolResultError("guidance_id is required"), nil
	}

	guidance, exists := g.layer1Guidance[guidanceID]
	if !exists {
		// Try to retrieve from storage
		if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(1, guidanceID); err == nil {
				if gd, ok := retrieved.(*layer1.GuidanceDocument); ok {
					guidance = gd
					// Update in-memory cache
					g.layer1Guidance[guidanceID] = guidance
					exists = true
				}
			}
		}
		if !exists {
			return mcp.NewToolResultErrorf("Guidance with ID '%s' not found. Use list_layer1_guidance to see available guidance.", guidanceID), nil
		}
	}

	outputFormat := request.GetString("output_format", "yaml")
	output, err := marshalOutput(guidance, outputFormat)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

// handleLoadLayer1FromFile loads a Layer 1 Guidance document from a file using Gemara library
func (g *GemaraAuthoringTools) handleLoadLayer1FromFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	guidance := &layer1.GuidanceDocument{}
	if err := guidance.LoadFile(filePath); err != nil {
		return mcp.NewToolResultErrorf("Failed to load Layer 1 Guidance from file: %v", err), nil
	}

	if guidance.Metadata.Id == "" {
		return mcp.NewToolResultError("Loaded guidance document missing metadata.id"), nil
	}

	// Store the loaded guidance in Gemara types storage
	g.layer1Guidance[guidance.Metadata.Id] = guidance
	
	// Also store to disk if storage is available
	if g.storage != nil {
		if err := g.storage.Add(1, guidance.Metadata.Id, guidance); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to store Layer 1 guidance to disk: %v\n", err)
		}
	}

	result := fmt.Sprintf("Successfully loaded Layer 1 Guidance:\n")
	result += fmt.Sprintf("- ID: %s\n", guidance.Metadata.Id)
	result += fmt.Sprintf("- Title: %s\n", guidance.Metadata.Title)
	result += fmt.Sprintf("- Description: %s\n", guidance.Metadata.Description)
	if guidance.Metadata.Version != "" {
		result += fmt.Sprintf("- Version: %s\n", guidance.Metadata.Version)
	}
	result += fmt.Sprintf("\nUse get_layer1_guidance with ID '%s' to retrieve full details.\n", guidance.Metadata.Id)

	return mcp.NewToolResultText(result), nil
}

// handleStoreLayer1YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 1 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer1YAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Validate with CUE first
	validationResult := g.PerformCUEValidation(yamlContent, 1)
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

	storedID, err := g.storage.StoreRawYAML(1, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	// Optionally load into memory for querying (lazy loading)
	// We'll load it on-demand when queried

	result := fmt.Sprintf("Successfully stored and validated Layer 1 Guidance:\n")
	result += fmt.Sprintf("- ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nUse get_layer1_guidance with ID '%s' to retrieve full details.\n", storedID)
	result += fmt.Sprintf("Use list_layer1_guidance to see all available guidance documents.\n")

	return mcp.NewToolResultText(result), nil
}

// handleSearchLayer1Guidance searches Layer 1 Guidance documents by name or description
func (g *GemaraAuthoringTools) handleSearchLayer1Guidance(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm := request.GetString("search_term", "")
	outputFormat := request.GetString("output_format", "yaml")

	if searchTerm == "" {
		return mcp.NewToolResultError("search_term is required"), nil
	}

	// Get all Layer 1 guidance entries
	var entries []*ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(1)
	} else {
		// Fallback to in-memory cache
		for guidanceID, guidance := range g.layer1Guidance {
			entries = append(entries, &ArtifactIndexEntry{
				ID:    guidanceID,
				Layer: 1,
				Title: guidance.Metadata.Title,
			})
		}
	}

	// Search through entries
	var matches []*layer1.GuidanceDocument
	searchLower := strings.ToLower(searchTerm)

	for _, entry := range entries {
		// Get full guidance document
		var guidance *layer1.GuidanceDocument
		if gd, exists := g.layer1Guidance[entry.ID]; exists {
			guidance = gd
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
				if gd, ok := retrieved.(*layer1.GuidanceDocument); ok {
					guidance = gd
					// Update cache
					g.layer1Guidance[entry.ID] = guidance
				}
			}
		}

		if guidance == nil {
			continue
		}

		// Search in title, description, and author
		titleMatch := strings.Contains(strings.ToLower(guidance.Metadata.Title), searchLower)
		descMatch := strings.Contains(strings.ToLower(guidance.Metadata.Description), searchLower)
		authorMatch := strings.Contains(strings.ToLower(guidance.Metadata.Author), searchLower)

		if titleMatch || descMatch || authorMatch {
			matches = append(matches, guidance)
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No Layer 1 Guidance documents found matching '%s'.\n\nUse list_layer1_guidance to see all available guidance documents.", searchTerm)), nil
	}

	// Format results
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(matches, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal results: %v", err), nil
		}
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// YAML format (default)
	result := fmt.Sprintf("# Search Results for '%s'\n\n", searchTerm)
	result += fmt.Sprintf("Found %d matching guidance document(s):\n\n", len(matches))

	for _, guidance := range matches {
		result += fmt.Sprintf("## %s\n", guidance.Metadata.Title)
		result += fmt.Sprintf("- **ID**: `%s`\n", guidance.Metadata.Id)
		if guidance.Metadata.Description != "" {
			result += fmt.Sprintf("- **Description**: %s\n", guidance.Metadata.Description)
		}
		if guidance.Metadata.Author != "" {
			result += fmt.Sprintf("- **Author**: %s\n", guidance.Metadata.Author)
		}
		result += "\n"
	}

	result += fmt.Sprintf("\nUse `get_layer1_guidance` with a guidance_id to get full details.\n")

	return mcp.NewToolResultText(result), nil
}

// REMOVED: handleCreateLayer1FromStructure - Use store_layer1_yaml instead
