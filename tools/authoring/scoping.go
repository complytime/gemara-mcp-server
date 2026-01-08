package authoring

import (
	"context"
	"fmt"
	"strings"

	"github.com/complytime/gemara-mcp-server/internal/consts"
	"github.com/complytime/gemara-mcp-server/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara"
)

// handleFindApplicableArtifacts finds Layer 1 and Layer 2 artifacts applicable to a given policy scope
// Uses storage index for efficient artifact discovery
func (g *GemaraAuthoringTools) handleFindApplicableArtifacts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract scope parameters
	boundaries := g.extractStringArray(request, "boundaries")
	technologies := g.extractStringArray(request, "technologies")
	providers := g.extractStringArray(request, "providers")
	outputFormat := request.GetString("output_format", "yaml")

	// Find applicable Layer 1 Guidance documents
	// Use storage index to get all Layer 1 artifacts, then load and check applicability
	var applicableLayer1 []string
	layer1Entries := g.getLayerEntries(consts.Layer1)
	for _, entry := range layer1Entries {
		guidance := g.loadLayer1Guidance(entry.ID)
		if guidance != nil && g.matchesLayer1Applicability(guidance, technologies, boundaries, providers) {
			applicableLayer1 = append(applicableLayer1, entry.ID)
		}
	}

	// Find applicable Layer 2 Controls
	type controlRef struct {
		catalogID string
		familyID  string
		controlID string
	}
	var applicableLayer2 []controlRef

	// Use storage index to get all Layer 2 catalogs
	layer2Entries := g.getLayerEntries(consts.Layer2)
	for _, entry := range layer2Entries {
		catalog := g.loadLayer2Catalog(entry.ID)

		if catalog == nil {
			continue
		}

		// Controls are now at catalog level, not nested in families
		for _, control := range catalog.Controls {
			if g.matchesLayer2Applicability(control, technologies, boundaries, providers) {
				applicableLayer2 = append(applicableLayer2, controlRef{
					catalogID: entry.ID,
					familyID:  control.Family,
					controlID: control.Id,
				})
			}
		}
	}

	// Format output
	if outputFormat == "json" {
		result := map[string]interface{}{
			"scope": map[string]interface{}{
				"boundaries":   boundaries,
				"technologies": technologies,
				"providers":    providers,
			},
			"layer1_guidance": applicableLayer1,
			"layer2_controls": make([]map[string]string, len(applicableLayer2)),
		}
		for i, ctrl := range applicableLayer2 {
			result["layer2_controls"].([]map[string]string)[i] = map[string]string{
				"catalog_id": ctrl.catalogID,
				"family_id":  ctrl.familyID,
				"control_id": ctrl.controlID,
			}
		}
		jsonBytes, err := marshalOutput(result, "json")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(jsonBytes), nil
	}

	// YAML format (default)
	var result strings.Builder
	result.WriteString("# Applicable Artifacts for Policy Scope\n\n")

	// Scope summary
	result.WriteString("## Policy Scope\n\n")
	if len(boundaries) > 0 {
		result.WriteString(fmt.Sprintf("- **Boundaries**: %s\n", strings.Join(boundaries, ", ")))
	}
	if len(technologies) > 0 {
		result.WriteString(fmt.Sprintf("- **Technologies**: %s\n", strings.Join(technologies, ", ")))
	}
	if len(providers) > 0 {
		result.WriteString(fmt.Sprintf("- **Providers**: %s\n", strings.Join(providers, ", ")))
	}
	if len(boundaries) == 0 && len(technologies) == 0 && len(providers) == 0 {
		result.WriteString("- **Scope**: All artifacts (no filters applied)\n")
	}
	result.WriteString("\n")

	// Layer 1 Guidance
	result.WriteString("## Layer 1 Guidance Documents\n\n")
	if len(applicableLayer1) == 0 {
		result.WriteString("No applicable Layer 1 Guidance documents found.\n\n")
	} else {
		result.WriteString(fmt.Sprintf("Found %d applicable guidance document(s):\n\n", len(applicableLayer1)))
		for _, guidanceID := range applicableLayer1 {
			guidance := g.layer1Guidance[guidanceID]
			result.WriteString(fmt.Sprintf("- **%s**: %s", guidanceID, guidance.Title))
			if guidance.Metadata.Version != "" {
				result.WriteString(fmt.Sprintf(" (v%s)", guidance.Metadata.Version))
			}
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	// Layer 2 Controls
	result.WriteString("## Layer 2 Controls\n\n")
	if len(applicableLayer2) == 0 {
		result.WriteString("No applicable Layer 2 Controls found.\n\n")
	} else {
		result.WriteString(fmt.Sprintf("Found %d applicable control(s):\n\n", len(applicableLayer2)))

		// Group by catalog
		catalogMap := make(map[string][]controlRef)
		for _, ctrl := range applicableLayer2 {
			catalogMap[ctrl.catalogID] = append(catalogMap[ctrl.catalogID], ctrl)
		}

		for catalogID, controls := range catalogMap {
			catalog := g.layer2Catalogs[catalogID]
			result.WriteString(fmt.Sprintf("### Catalog: %s\n\n", catalog.Title))
			for _, ctrl := range controls {
				control := g.findControl(catalogID, ctrl.familyID, ctrl.controlID)
				if control != nil {
					result.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", ctrl.controlID, ctrl.familyID, control.Title))
				} else {
					result.WriteString(fmt.Sprintf("- **%s** (%s)\n", ctrl.controlID, ctrl.familyID))
				}
			}
			result.WriteString("\n")
		}
	}

	// Summary
	result.WriteString("## Summary\n\n")
	result.WriteString(fmt.Sprintf("- **Layer 1 Guidance**: %d document(s)\n", len(applicableLayer1)))
	result.WriteString(fmt.Sprintf("- **Layer 2 Controls**: %d control(s)\n", len(applicableLayer2)))
	result.WriteString(fmt.Sprintf("- **Total Applicable Artifacts**: %d\n", len(applicableLayer1)+len(applicableLayer2)))

	return mcp.NewToolResultText(result.String()), nil
}

// getLayerEntries gets entries for a given layer from storage or cache
func (g *GemaraAuthoringTools) getLayerEntries(layer int) []*storage.ArtifactIndexEntry {
	if g.storage != nil {
		return g.storage.List(layer)
	}

	// Fallback to in-memory cache
	var entries []*storage.ArtifactIndexEntry
	switch layer {
	case consts.Layer1:
		for guidanceID, guidance := range g.layer1Guidance {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    guidanceID,
				Layer: consts.Layer1,
				Title: guidance.Title,
			})
		}
	case consts.Layer2:
		for catalogID, catalog := range g.layer2Catalogs {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    catalogID,
				Layer: consts.Layer2,
				Title: catalog.Title,
			})
		}
	}
	return entries
}

// loadLayer1Guidance loads a Layer 1 Guidance document from cache or storage
func (g *GemaraAuthoringTools) loadLayer1Guidance(guidanceID string) *gemara.GuidanceDocument {
	if gd, exists := g.layer1Guidance[guidanceID]; exists {
		return gd
	}
	if g.storage != nil {
		if retrieved, err := g.storage.Retrieve(consts.Layer1, guidanceID); err == nil {
			if gd, ok := retrieved.(*gemara.GuidanceDocument); ok {
				g.layer1Guidance[guidanceID] = gd
				return gd
			}
		}
	}
	return nil
}

// loadLayer2Catalog loads a Layer 2 Catalog from cache or storage
func (g *GemaraAuthoringTools) loadLayer2Catalog(catalogID string) *gemara.Catalog {
	if c, exists := g.layer2Catalogs[catalogID]; exists {
		return c
	}
	if g.storage != nil {
		if retrieved, err := g.storage.Retrieve(consts.Layer2, catalogID); err == nil {
			if c, ok := retrieved.(*gemara.Catalog); ok {
				g.layer2Catalogs[catalogID] = c
				return c
			}
		}
	}
	return nil
}

// findControlFamily finds a control family by ID
func (g *GemaraAuthoringTools) findControlFamily(catalogID, familyID string) *gemara.Family {
	catalog, ok := g.layer2Catalogs[catalogID]
	if !ok {
		return nil
	}
	for _, family := range catalog.Families {
		if family.Id == familyID {
			return &family
		}
	}
	return nil
}

// findControl finds a control by ID
func (g *GemaraAuthoringTools) findControl(catalogID, familyID, controlID string) *gemara.Control {
	catalog, ok := g.layer2Catalogs[catalogID]
	if !ok {
		return nil
	}
	for i := range catalog.Controls {
		if catalog.Controls[i].Id == controlID && catalog.Controls[i].Family == familyID {
			return &catalog.Controls[i]
		}
	}
	return nil
}
