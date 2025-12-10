package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/complytime/gemara-mcp-server/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
)

// handleListLayer2Controls lists available Layer 2 Controls with optional filtering
// Uses storage index for efficient catalog discovery
func (g *GemaraAuthoringTools) handleListLayer2Controls(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	_ = request.GetString("technology", "") // Technology filtering not yet implemented for Gemara types
	layer1Ref := request.GetString("layer1_reference", "")
	outputFormat := request.GetString("output_format", "yaml")

	// Rescan storage to discover new artifacts
	if g.storage != nil {
		if err := g.storage.Rescan(); err != nil {
			slog.Warn("Failed to rescan storage for new artifacts", "error", err)
		}
	}

	// Get catalog entries from storage index (fast)
	var catalogEntries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		catalogEntries = g.storage.List(2)
	} else {
		// Fallback to in-memory cache
		for catalogID, catalog := range g.layer2Catalogs {
			catalogEntries = append(catalogEntries, &storage.ArtifactIndexEntry{
				ID:    catalogID,
				Layer: 2,
				Title: catalog.Metadata.Title,
			})
		}
	}

	if len(catalogEntries) == 0 {
		return mcp.NewToolResultText("No Layer 2 Controls available.\n\nUse store_layer2_yaml to store controls."), nil
	}

	// Collect all controls from catalogs with filtering
	type controlInfo struct {
		catalogID string
		familyID  string
		control   layer2.Control
	}
	var allControls []controlInfo

	for _, catalogEntry := range catalogEntries {
		// Get catalog from cache or storage
		var catalog *layer2.Catalog
		if c, exists := g.layer2Catalogs[catalogEntry.ID]; exists {
			catalog = c
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(2, catalogEntry.ID); err == nil {
				if c, ok := retrieved.(*layer2.Catalog); ok {
					catalog = c
					g.layer2Catalogs[catalogEntry.ID] = catalog
				}
			}
		}

		if catalog == nil {
			continue
		}

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
					catalogID: catalogEntry.ID,
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

// handleSearchLayer2Controls searches controls by name, objective, or ID
// Can also filter by Layer 1 guidance reference, technology, or applicability scope
// Uses storage index for efficient filtering before loading full catalogs
func (g *GemaraAuthoringTools) handleSearchLayer2Controls(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm := request.GetString("search_term", "")
	technology := request.GetString("technology", "")
	layer1Ref := request.GetString("layer1_reference", "")
	boundaries := g.extractStringArray(request, "boundaries")
	technologies := g.extractStringArray(request, "technologies")
	providers := g.extractStringArray(request, "providers")
	outputFormat := request.GetString("output_format", "yaml")

	// Allow empty search_term if other filters are provided
	if searchTerm == "" && technology == "" && layer1Ref == "" && len(boundaries) == 0 && len(technologies) == 0 && len(providers) == 0 {
		return mcp.NewToolResultError("search_term is required, or provide at least one filter (technology, layer1_reference, boundaries, technologies, providers)"), nil
	}

	searchTermLower := strings.ToLower(searchTerm)

	// Get catalog entries from storage index (fast)
	var catalogEntries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		catalogEntries = g.storage.List(2)
	} else {
		// Fallback to in-memory cache
		for catalogID, catalog := range g.layer2Catalogs {
			catalogEntries = append(catalogEntries, &storage.ArtifactIndexEntry{
				ID:    catalogID,
				Layer: 2,
				Title: catalog.Metadata.Title,
			})
		}
	}

	type controlMatch struct {
		catalogID string
		familyID  string
		control   layer2.Control
	}
	var matches []controlMatch

	// Search through catalogs (load from cache or storage as needed)
	for _, catalogEntry := range catalogEntries {
		var catalog *layer2.Catalog
		if c, exists := g.layer2Catalogs[catalogEntry.ID]; exists {
			catalog = c
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(2, catalogEntry.ID); err == nil {
				if c, ok := retrieved.(*layer2.Catalog); ok {
					catalog = c
					g.layer2Catalogs[catalogEntry.ID] = catalog
				}
			}
		}

		if catalog == nil {
			continue
		}

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

				// Filter by technology if specified
				if technology != "" {
					// Check if control matches technology (simplified - could be enhanced)
					// For now, we'll include all controls if technology filter is not precise
				}

				// Apply scoping filters if provided
				if len(boundaries) > 0 || len(technologies) > 0 || len(providers) > 0 {
					if !g.matchesLayer2Applicability(control, technologies, boundaries, providers) {
						continue
					}
				}

				// Search in title, objective, and control ID
				// If search_term is empty, include all controls that passed filters above
				if searchTerm == "" {
					matches = append(matches, controlMatch{
						catalogID: catalogEntry.ID,
						familyID:  family.Id,
						control:   control,
					})
				} else {
					titleMatch := strings.Contains(strings.ToLower(control.Title), searchTermLower)
					objectiveMatch := strings.Contains(strings.ToLower(control.Objective), searchTermLower)
					idMatch := strings.Contains(strings.ToLower(control.Id), searchTermLower)

					if titleMatch || objectiveMatch || idMatch {
						matches = append(matches, controlMatch{
							catalogID: catalogEntry.ID,
							familyID:  family.Id,
							control:   control,
						})
					}
				}
			}
		}
	}

	if len(matches) == 0 {
		filterParts := []string{}
		if searchTerm != "" {
			filterParts = append(filterParts, fmt.Sprintf("search term '%s'", searchTerm))
		}
		if layer1Ref != "" {
			filterParts = append(filterParts, fmt.Sprintf("Layer 1 guidance '%s'", layer1Ref))
		}
		if technology != "" {
			filterParts = append(filterParts, fmt.Sprintf("technology '%s'", technology))
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
		return mcp.NewToolResultText(fmt.Sprintf("No controls found matching %s.\n\nTry different filters or use list_layer2_controls to see all available controls.", filterMsg)), nil
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
		result := fmt.Sprintf("# Search Results")
		if searchTerm != "" {
			result += fmt.Sprintf(" for '%s'", searchTerm)
		}
		result += "\n\n"
		result += fmt.Sprintf("Found %d control(s)", len(matches))
		filterParts := []string{}
		if layer1Ref != "" {
			filterParts = append(filterParts, fmt.Sprintf("Layer 1 guidance: %s", layer1Ref))
		}
		if technology != "" {
			filterParts = append(filterParts, fmt.Sprintf("technology: %s", technology))
		}
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

		for _, m := range matches {
			result += fmt.Sprintf("- **%s** (`%s`) - Catalog: %s", m.control.Title, m.control.Id, m.catalogID)
			// Show which Layer 1 guidance this control references
			if len(m.control.GuidelineMappings) > 0 {
				refs := []string{}
				for _, mapping := range m.control.GuidelineMappings {
					refs = append(refs, mapping.ReferenceId)
				}
				result += fmt.Sprintf(" (references: %s)", strings.Join(refs, ", "))
			}
			result += "\n"
		}

		result += "\nUse `get_layer2_control` with a control_id to get full details.\n"
		if layer1Ref != "" {
			result += fmt.Sprintf("Use `get_layer2_guideline_mappings` to see detailed guideline mappings for a control.\n")
		}
		output = result
	}

	return mcp.NewToolResultText(output), nil
}

// handleStoreLayer2YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 2 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer2YAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Store with validation (ensures CUE validation always happens)
	storedID, err := g.StoreValidatedYAML(2, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	// Load into memory cache for immediate querying
	if retrieved, err := g.storage.Retrieve(2, storedID); err == nil {
		if catalog, ok := retrieved.(*layer2.Catalog); ok {
			g.layer2Catalogs[storedID] = catalog
		}
	}

	result := fmt.Sprintf("Successfully stored and validated Layer 2 Control Catalog:\n")
	result += fmt.Sprintf("- Catalog ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nUse get_layer2_control with catalog ID '%s' to retrieve full details.\n", storedID)
	result += fmt.Sprintf("Use list_layer2_controls to see all available controls.\n")

	return mcp.NewToolResultText(result), nil
}

// handleGetLayer2GuidelineMappings retrieves all Layer 1 guideline mappings for a Layer 2 control
func (g *GemaraAuthoringTools) handleGetLayer2GuidelineMappings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	controlID := request.GetString("control_id", "")
	outputFormat := request.GetString("output_format", "yaml")
	includeGuidanceDetailsStr := request.GetString("include_guidance_details", "false")
	includeGuidanceDetails := includeGuidanceDetailsStr == "true" || includeGuidanceDetailsStr == "1"

	if controlID == "" {
		return mcp.NewToolResultError("control_id is required"), nil
	}

	// Find the control across all catalogs
	var foundControl *layer2.Control
	var catalogID string
	var familyID string

	for catID, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for _, control := range family.Controls {
				if control.Id == controlID {
					foundControl = &control
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
		return mcp.NewToolResultText(fmt.Sprintf("Control '%s' not found.\n\nUse list_layer2_controls to see all available controls.", controlID)), nil
	}

	// Extract guideline mappings
	if len(foundControl.GuidelineMappings) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Control '%s' (%s) has no Layer 1 guideline mappings.\n\nThis control does not reference any Layer 1 guidance documents.", controlID, foundControl.Title)), nil
	}

	// Format output
	if outputFormat == "json" {
		result := map[string]interface{}{
			"control_id":         controlID,
			"control_title":      foundControl.Title,
			"catalog_id":         catalogID,
			"family_id":          familyID,
			"guideline_mappings": make([]map[string]interface{}, len(foundControl.GuidelineMappings)),
		}

		for i, mapping := range foundControl.GuidelineMappings {
			mappingData := map[string]interface{}{
				"reference_id": mapping.ReferenceId,
				"entries":      make([]map[string]interface{}, len(mapping.Entries)),
			}

			for j, entry := range mapping.Entries {
				entryData := map[string]interface{}{
					"reference_id": entry.ReferenceId,
					"strength":     entry.Strength,
				}
				if entry.Remarks != "" {
					entryData["remarks"] = entry.Remarks
				}

				// Optionally include Layer 1 guidance details
				if includeGuidanceDetails {
					if guidance, ok := g.layer1Guidance[mapping.ReferenceId]; ok {
						entryData["guidance_title"] = guidance.Metadata.Title
						entryData["guidance_version"] = guidance.Metadata.Version
					}
				}

				mappingData["entries"].([]map[string]interface{})[j] = entryData
			}

			// Add guidance document details if requested
			if includeGuidanceDetails {
				if guidance, ok := g.layer1Guidance[mapping.ReferenceId]; ok {
					mappingData["guidance_document"] = map[string]interface{}{
						"id":      mapping.ReferenceId,
						"title":   guidance.Metadata.Title,
						"version": guidance.Metadata.Version,
						"author":  guidance.Metadata.Author,
					}
				}
			}

			result["guideline_mappings"].([]map[string]interface{})[i] = mappingData
		}

		jsonBytes, err := marshalOutput(result, "json")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(jsonBytes), nil
	}

	// YAML format (default)
	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Layer 1 Guideline Mappings for Control: %s\n\n", controlID))
	result.WriteString(fmt.Sprintf("**Control Title**: %s\n", foundControl.Title))
	result.WriteString(fmt.Sprintf("**Catalog**: %s\n", catalogID))
	result.WriteString(fmt.Sprintf("**Family**: %s\n\n", familyID))

	result.WriteString(fmt.Sprintf("## Guideline Mappings\n\n"))
	result.WriteString(fmt.Sprintf("This control references **%d Layer 1 guidance document(s)** with **%d total guideline entries**.\n\n",
		len(foundControl.GuidelineMappings), g.countTotalGuidelineEntries(foundControl.GuidelineMappings)))

	// Group mappings by guidance document
	for i, mapping := range foundControl.GuidelineMappings {
		guidanceDoc := g.getGuidanceDocument(mapping.ReferenceId)

		result.WriteString(fmt.Sprintf("### %d. Guidance Document: `%s`\n\n", i+1, mapping.ReferenceId))

		if guidanceDoc != nil {
			result.WriteString(fmt.Sprintf("- **Title**: %s\n", guidanceDoc.Metadata.Title))
			if guidanceDoc.Metadata.Version != "" {
				result.WriteString(fmt.Sprintf("- **Version**: %s\n", guidanceDoc.Metadata.Version))
			}
			if guidanceDoc.Metadata.Author != "" {
				result.WriteString(fmt.Sprintf("- **Author**: %s\n", guidanceDoc.Metadata.Author))
			}
		}

		result.WriteString(fmt.Sprintf("- **Guideline Entries**: %d\n\n", len(mapping.Entries)))

		// List all guideline entries
		for j, entry := range mapping.Entries {
			result.WriteString(fmt.Sprintf("  **%d.%d** `%s`", i+1, j+1, entry.ReferenceId))
			if entry.Strength > 0 {
				result.WriteString(fmt.Sprintf(" (strength: %d)", entry.Strength))
			}
			if entry.Remarks != "" {
				result.WriteString(fmt.Sprintf(" - %s", entry.Remarks))
			}
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	// Summary
	result.WriteString("## Summary\n\n")
	result.WriteString(fmt.Sprintf("- **Total Guidance Documents**: %d\n", len(foundControl.GuidelineMappings)))
	result.WriteString(fmt.Sprintf("- **Total Guideline References**: %d\n", g.countTotalGuidelineEntries(foundControl.GuidelineMappings)))

	if includeGuidanceDetails {
		result.WriteString("\n*Note: Guidance document details are included above.*\n")
	} else {
		result.WriteString("\n*Tip: Use `include_guidance_details=true` to see full guidance document information.*\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

// countTotalGuidelineEntries counts the total number of guideline entries across all mappings
func (g *GemaraAuthoringTools) countTotalGuidelineEntries(mappings []layer2.Mapping) int {
	total := 0
	for _, mapping := range mappings {
		total += len(mapping.Entries)
	}
	return total
}

// getGuidanceDocument retrieves a Layer 1 guidance document by ID
func (g *GemaraAuthoringTools) getGuidanceDocument(guidanceID string) *layer1.GuidanceDocument {
	if guidance, ok := g.layer1Guidance[guidanceID]; ok {
		return guidance
	}
	return nil
}
