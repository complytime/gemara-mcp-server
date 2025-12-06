package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
)

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
			"control_id":   controlID,
			"control_title": foundControl.Title,
			"catalog_id":    catalogID,
			"family_id":     familyID,
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
