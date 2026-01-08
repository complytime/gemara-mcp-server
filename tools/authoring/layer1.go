package authoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/complytime/gemara-mcp-server/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara"
)

// handleListLayer1Guidance lists all available Layer 1 Guidance documents
func (g *GemaraAuthoringTools) handleListLayer1Guidance(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Rescan storage to discover new artifacts
	if g.storage != nil {
		if err := g.storage.Rescan(); err != nil {
			slog.Warn("Failed to rescan storage for new artifacts", "error", err)
		}
	}

	// Get list from storage if available, otherwise use in-memory cache
	var entries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(1)
	} else {
		// Fallback to in-memory cache
		for guidanceID, guidance := range g.layer1Guidance {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    guidanceID,
				Layer: 1,
				Title: guidance.Title,
			})
		}
	}

	totalCount := len(entries)
	if totalCount == 0 {
		return mcp.NewToolResultText("No Layer 1 Guidance documents available.\n\nUse store_layer1_yaml to store guidance documents."), nil
	}

	result := fmt.Sprintf("# Available Layer 1 Guidance Documents\n\n")
	result += fmt.Sprintf("Total: %d guidance document(s)\n\n", totalCount)

	for _, entry := range entries {
		// Try to get full details from cache or storage
		var guidance *gemara.GuidanceDocument
		if gd, exists := g.layer1Guidance[entry.ID]; exists {
			guidance = gd
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
				if gd, ok := retrieved.(*gemara.GuidanceDocument); ok {
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
			if guidance.Metadata.Author.Name != "" {
				result += fmt.Sprintf("- **Author**: %s\n", guidance.Metadata.Author.Name)
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
func (g *GemaraAuthoringTools) handleGetLayer1Guidance(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				if gd, ok := retrieved.(*gemara.GuidanceDocument); ok {
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

// handleStoreLayer1YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 1 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer1YAML(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Store with validation (ensures CUE validation always happens)
	storedID, err := g.StoreValidatedYAML(1, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	// Load into memory cache for immediate querying
	if retrieved, err := g.storage.Retrieve(1, storedID); err == nil {
		if guidance, ok := retrieved.(*gemara.GuidanceDocument); ok {
			g.layer1Guidance[storedID] = guidance
		}
	}

	result := fmt.Sprintf("Successfully stored and validated Layer 1 Guidance:\n")
	result += fmt.Sprintf("- ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nUse get_layer1_guidance with ID '%s' to retrieve full details.\n", storedID)
	result += fmt.Sprintf("Use list_layer1_guidance to see all available guidance documents.\n")

	return mcp.NewToolResultText(result), nil
}

// handleSearchLayer1Guidance searches Layer 1 Guidance documents by name, description, or author
// Can optionally filter by applicability scope (boundaries, technologies, providers)
// Uses storage index for efficient filtering before loading full documents
func (g *GemaraAuthoringTools) handleSearchLayer1Guidance(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm := request.GetString("search_term", "")
	boundaries := g.extractStringArray(request, "boundaries")
	technologies := g.extractStringArray(request, "technologies")
	providers := g.extractStringArray(request, "providers")
	outputFormat := request.GetString("output_format", "yaml")

	// Allow empty search_term if scoping filters are provided
	if searchTerm == "" && len(boundaries) == 0 && len(technologies) == 0 && len(providers) == 0 {
		return mcp.NewToolResultError("search_term is required, or provide at least one scoping filter (boundaries, technologies, providers)"), nil
	}

	// Get all Layer 1 guidance entries from storage index (fast)
	var entries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(1)
	} else {
		// Fallback to in-memory cache
		for guidanceID, guidance := range g.layer1Guidance {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    guidanceID,
				Layer: 1,
				Title: guidance.Title,
			})
		}
	}

	// First pass: filter by title match in index (fast, no need to load full documents)
	// If search_term is empty, use all entries (scoping-only search)
	searchLower := strings.ToLower(searchTerm)
	var candidateEntries []*storage.ArtifactIndexEntry
	if searchTerm == "" {
		// Scoping-only: use all entries
		candidateEntries = entries
	} else {
		// Text search: filter by title or ID from index
		for _, entry := range entries {
			if strings.Contains(strings.ToLower(entry.Title), searchLower) ||
				strings.Contains(strings.ToLower(entry.ID), searchLower) {
				candidateEntries = append(candidateEntries, entry)
			}
		}
	}

	// Second pass: load full documents only for candidates and check description/author
	var matches []*gemara.GuidanceDocument
	for _, entry := range candidateEntries {
		// Get full guidance document (from cache or storage)
		var guidance *gemara.GuidanceDocument
		if gd, exists := g.layer1Guidance[entry.ID]; exists {
			guidance = gd
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
				if gd, ok := retrieved.(*gemara.GuidanceDocument); ok {
					guidance = gd
					// Update cache
					g.layer1Guidance[entry.ID] = guidance
				}
			}
		}

		if guidance == nil {
			continue
		}

		// Apply scoping filters if provided
		if len(boundaries) > 0 || len(technologies) > 0 || len(providers) > 0 {
			if !g.matchesLayer1Applicability(guidance, technologies, boundaries, providers) {
				continue
			}
		}

		// If search_term is empty (scoping-only), include all that passed scoping
		// Otherwise, already matched on title/ID in first pass
		if searchTerm == "" {
			matches = append(matches, guidance)
		} else {
			// Already matched on title/ID in first pass, include it
			matches = append(matches, guidance)
		}
	}

	// If no matches from index, and we have a search term, do a full search (slower but more thorough)
	if len(matches) == 0 && searchTerm != "" {
		for _, entry := range entries {
			var guidance *gemara.GuidanceDocument
			if gd, exists := g.layer1Guidance[entry.ID]; exists {
				guidance = gd
			} else if g.storage != nil {
				if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
					if gd, ok := retrieved.(*gemara.GuidanceDocument); ok {
						guidance = gd
						g.layer1Guidance[entry.ID] = guidance
					}
				}
			}

			if guidance == nil {
				continue
			}

			// Full text search
			titleMatch := strings.Contains(strings.ToLower(guidance.Title), searchLower)
			descMatch := strings.Contains(strings.ToLower(guidance.Metadata.Description), searchLower)
			authorMatch := strings.Contains(strings.ToLower(guidance.Metadata.Author.Name), searchLower)

			if titleMatch || descMatch || authorMatch {
				// Apply scoping filters if provided
				if len(boundaries) > 0 || len(technologies) > 0 || len(providers) > 0 {
					if !g.matchesLayer1Applicability(guidance, technologies, boundaries, providers) {
						continue
					}
				}
				matches = append(matches, guidance)
			}
		}
	}

	if len(matches) == 0 {
		var filterParts []string
		if searchTerm != "" {
			filterParts = append(filterParts, fmt.Sprintf("search term '%s'", searchTerm))
		}
		if len(boundaries) > 0 {
			filterParts = append(filterParts, fmt.Sprintf("boundaries %v", boundaries))
		}
		if len(technologies) > 0 {
			filterParts = append(filterParts, fmt.Sprintf("technologies %v", technologies))
		}
		if len(providers) > 0 {
			filterParts = append(filterParts, fmt.Sprintf("providers %v", providers))
		}
		filterMsg := strings.Join(filterParts, ", ")
		return mcp.NewToolResultText(fmt.Sprintf("No Layer 1 Guidance documents found matching %s.\n\nUse list_layer1_guidance to see all available guidance documents.", filterMsg)), nil
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
	result := "# Search Results"
	if searchTerm != "" {
		result += fmt.Sprintf(" for '%s'", searchTerm)
	}
	result += "\n\n"
	result += fmt.Sprintf("Found %d matching guidance document(s)", len(matches))
	var filterParts []string
	if len(boundaries) > 0 {
		filterParts = append(filterParts, fmt.Sprintf("boundaries: %v", boundaries))
	}
	if len(technologies) > 0 {
		filterParts = append(filterParts, fmt.Sprintf("technologies: %v", technologies))
	}
	if len(providers) > 0 {
		filterParts = append(filterParts, fmt.Sprintf("providers: %v", providers))
	}
	if len(filterParts) > 0 {
		result += fmt.Sprintf(" (filtered by %s)", strings.Join(filterParts, ", "))
	}
	result += "\n\n"

	for _, guidance := range matches {
		result += fmt.Sprintf("## %s\n", guidance.Title)
		result += fmt.Sprintf("- **ID**: `%s`\n", guidance.Metadata.Id)
		if guidance.Metadata.Description != "" {
			result += fmt.Sprintf("- **Description**: %s\n", guidance.Metadata.Description)
		}
		if guidance.Metadata.Author.Name != "" {
			result += fmt.Sprintf("- **Author**: %s\n", guidance.Metadata.Author.Name)
		}
		result += "\n"
	}

	result += fmt.Sprintf("\nUse `get_layer1_guidance` with a guidance_id to get full details.\n")

	return mcp.NewToolResultText(result), nil
}
