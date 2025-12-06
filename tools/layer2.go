package tools

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer2"
)

// REMOVED: handleCreateLayer2Control - Use store_layer2_yaml instead

// handleListLayer2Controls lists available Layer 2 Controls with optional filtering
func (g *GemaraAuthoringTools) handleListLayer2Controls(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	_ = request.GetString("technology", "") // Technology filtering not yet implemented for Gemara types
	layer1Ref := request.GetString("layer1_reference", "")
	outputFormat := request.GetString("output_format", "yaml")

	totalCatalogs := len(g.layer2Catalogs)
	if totalCatalogs == 0 {
		return mcp.NewToolResultText("No Layer 2 Controls available.\n\nUse store_layer2_yaml to store controls or load_layer2_from_file to load from disk."), nil
	}

	// Collect all controls from catalogs with filtering
	type controlInfo struct {
		catalogID string
		familyID  string
		control   layer2.Control
	}
	var allControls []controlInfo

	for catalogID, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for _, control := range family.Controls {
				// Filter by layer1_reference if specified
				if layer1Ref != "" {
					found := false
					for _, mapping := range control.GuidelineMappings {
						if mapping.ReferenceId == layer1Ref {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				allControls = append(allControls, controlInfo{
					catalogID: catalogID,
					familyID:  family.Id,
					control:   control,
				})
			}
		}
	}

	if len(allControls) == 0 {
		filterMsg := ""
		if layer1Ref != "" {
			filterMsg += fmt.Sprintf(" referencing Layer 1 guidance '%s'", layer1Ref)
		}
		return mcp.NewToolResultText(fmt.Sprintf("No Layer 2 Controls found%s.\n\nTry removing filters or use store_layer2_yaml to store new controls.", filterMsg)), nil
	}

	var output string
	if outputFormat == "json" {
		// Convert to JSON format
		controlsJSON := make([]map[string]interface{}, len(allControls))
		for i, ci := range allControls {
			controlsJSON[i] = map[string]interface{}{
				"control_id": ci.control.Id,
				"title":      ci.control.Title,
				"objective":  ci.control.Objective,
				"catalog_id": ci.catalogID,
				"family_id":  ci.familyID,
			}
		}
		output, err := marshalOutput(controlsJSON, outputFormat)
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(output), nil
	} else {
		result := fmt.Sprintf("# Available Layer 2 Controls\n\n")
		result += fmt.Sprintf("Total: %d control(s)", len(allControls))
		if layer1Ref != "" {
			result += fmt.Sprintf(" (filtered by Layer 1 reference: %s)", layer1Ref)
		}
		result += "\n\n"

		// Group by catalog
		catalogMap := make(map[string][]controlInfo)
		for _, ci := range allControls {
			catalogMap[ci.catalogID] = append(catalogMap[ci.catalogID], ci)
		}

		for catalogID, controls := range catalogMap {
			catalog := g.layer2Catalogs[catalogID]
			result += fmt.Sprintf("## Catalog: %s\n", catalog.Metadata.Title)
			result += fmt.Sprintf("- **Catalog ID**: `%s`\n", catalogID)
			if catalog.Metadata.Description != "" {
				result += fmt.Sprintf("- **Description**: %s\n", catalog.Metadata.Description)
			}
			result += fmt.Sprintf("- **Controls**: %d\n\n", len(controls))

			for _, ci := range controls {
				result += fmt.Sprintf("### %s (`%s`)\n", ci.control.Title, ci.control.Id)
				result += fmt.Sprintf("- **Objective**: %s\n", ci.control.Objective)
				if len(ci.control.GuidelineMappings) > 0 {
					result += fmt.Sprintf("- **References Layer 1**: ")
					for i, mapping := range ci.control.GuidelineMappings {
						if i > 0 {
							result += ", "
						}
						result += fmt.Sprintf("`%s`", mapping.ReferenceId)
					}
					result += "\n"
				}
				result += "\n"
			}
		}

		result += "\nUse `get_layer2_control` with a control_id to get full details.\n"
		result += "Use these control IDs in `layer2_controls` when creating Layer 3 policies.\n"
		output = result
	}

	return mcp.NewToolResultText(output), nil
}

// handleGetLayer2Control gets detailed information about a specific Layer 2 Control
func (g *GemaraAuthoringTools) handleGetLayer2Control(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	controlID := request.GetString("control_id", "")
	outputFormat := request.GetString("output_format", "yaml")

	if controlID == "" {
		return mcp.NewToolResultError("control_id is required"), nil
	}

	// Search for control in all catalogs
	var foundControl *layer2.Control
	var catalogID, familyID string
	for catID, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for i := range family.Controls {
				if family.Controls[i].Id == controlID {
					foundControl = &family.Controls[i]
					catalogID = catID
					familyID = family.Id
					break
				}
			}
			if foundControl != nil {
				break
			}
		}
		if foundControl != nil {
			break
		}
	}

	if foundControl == nil {
		return mcp.NewToolResultErrorf("Control with ID '%s' not found. Use list_layer2_controls to see available controls.", controlID), nil
	}

	controlOutput, err := marshalOutput(foundControl, outputFormat)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal: %v", err), nil
	}
	output := fmt.Sprintf("Catalog: %s\nFamily: %s\n\n%s", catalogID, familyID, controlOutput)

	return mcp.NewToolResultText(output), nil
}

// handleSearchLayer2Controls searches controls by name or description
func (g *GemaraAuthoringTools) handleSearchLayer2Controls(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm := request.GetString("search_term", "")
	technology := request.GetString("technology", "")
	outputFormat := request.GetString("output_format", "yaml")

	if searchTerm == "" {
		return mcp.NewToolResultError("search_term is required"), nil
	}

	searchTermLower := fmt.Sprintf("%s", searchTerm)
	// Search through all catalogs
	type controlMatch struct {
		catalogID string
		familyID  string
		control   layer2.Control
	}
	var matches []controlMatch

	for catalogID, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for _, control := range family.Controls {
				// Search in title and objective
				titleLower := fmt.Sprintf("%s", control.Title)
				objectiveLower := fmt.Sprintf("%s", control.Objective)
				if containsIgnoreCase(titleLower, searchTermLower) || containsIgnoreCase(objectiveLower, searchTermLower) {
					matches = append(matches, controlMatch{
						catalogID: catalogID,
						familyID:  family.Id,
						control:   control,
					})
				}
			}
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No controls found matching '%s'%s.\n\nTry a different search term or use list_layer2_controls to see all available controls.", searchTerm, func() string {
			if technology != "" {
				return fmt.Sprintf(" with technology '%s'", technology)
			}
			return ""
		}())), nil
	}

	var output string
	if outputFormat == "json" {
		matchesJSON := make([]map[string]interface{}, len(matches))
		for i, m := range matches {
			matchesJSON[i] = map[string]interface{}{
				"control_id": m.control.Id,
				"title":      m.control.Title,
				"objective":  m.control.Objective,
				"catalog_id": m.catalogID,
				"family_id":  m.familyID,
			}
		}
		output, err := marshalOutput(matchesJSON, outputFormat)
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(output), nil
	} else {
		result := fmt.Sprintf("# Search Results for '%s'\n\n", searchTerm)
		result += fmt.Sprintf("Found %d control(s)", len(matches))
		if technology != "" {
			result += fmt.Sprintf(" (filtered by technology: %s)", technology)
		}
		result += "\n\n"

		for _, m := range matches {
			result += fmt.Sprintf("- **%s** (`%s`) - Catalog: %s\n", m.control.Title, m.control.Id, m.catalogID)
		}

		result += "\nUse `get_layer2_control` with a control_id to get full details.\n"
		output = result
	}

	return mcp.NewToolResultText(output), nil
}

// handleLoadLayer2FromFile loads a Layer 2 Control Catalog from a file using Gemara library
func (g *GemaraAuthoringTools) handleLoadLayer2FromFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	catalog := &layer2.Catalog{}
	if err := catalog.LoadFile(filePath); err != nil {
		return mcp.NewToolResultErrorf("Failed to load Layer 2 Catalog from file: %v", err), nil
	}

	if catalog.Metadata.Id == "" {
		return mcp.NewToolResultError("Loaded catalog missing metadata.id"), nil
	}

	// Store the loaded catalog in Gemara types storage
	g.layer2Catalogs[catalog.Metadata.Id] = catalog

	// Count controls
	totalControls := 0
	for _, family := range catalog.ControlFamilies {
		totalControls += len(family.Controls)
	}

	result := fmt.Sprintf("Successfully loaded Layer 2 Control Catalog:\n")
	result += fmt.Sprintf("- Catalog ID: %s\n", catalog.Metadata.Id)
	result += fmt.Sprintf("- Title: %s\n", catalog.Metadata.Title)
	result += fmt.Sprintf("- Description: %s\n", catalog.Metadata.Description)
	result += fmt.Sprintf("- Control Families: %d\n", len(catalog.ControlFamilies))
	result += fmt.Sprintf("- Total Controls: %d\n", totalControls)
	result += fmt.Sprintf("\nUse list_layer2_controls to see all controls in this catalog.\n")

	return mcp.NewToolResultText(result), nil
}

// handleStoreLayer2YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 2 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer2YAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Validate with CUE first
	validationResult := g.PerformCUEValidation(yamlContent, 2)
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

	storedID, err := g.storage.StoreRawYAML(2, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	result := fmt.Sprintf("Successfully stored and validated Layer 2 Control Catalog:\n")
	result += fmt.Sprintf("- Catalog ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nUse get_layer2_control with catalog ID '%s' to retrieve full details.\n", storedID)
	result += fmt.Sprintf("Use list_layer2_controls to see all available controls.\n")

	return mcp.NewToolResultText(result), nil
}
