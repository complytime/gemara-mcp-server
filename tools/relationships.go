package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
)

// handleGetArtifactRelationships visualizes cross-layer dependencies for artifacts
func (g *GemaraAuthoringTools) handleGetArtifactRelationships(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artifactID := request.GetString("artifact_id", "")
	artifactType := request.GetString("artifact_type", "") // "layer1", "layer2", "layer3"
	outputFormat := request.GetString("output_format", "yaml")

	if artifactID == "" {
		return mcp.NewToolResultError("artifact_id is required"), nil
	}

	if artifactType == "" {
		return mcp.NewToolResultError("artifact_type is required (must be 'layer1', 'layer2', or 'layer3')"), nil
	}

	relationships := &ArtifactRelationships{
		ArtifactID:   artifactID,
		ArtifactType: artifactType,
	}

	var err error
	switch artifactType {
	case "layer1":
		err = g.buildLayer1Relationships(artifactID, relationships)
	case "layer2":
		err = g.buildLayer2Relationships(artifactID, relationships)
	case "layer3":
		err = g.buildLayer3Relationships(artifactID, relationships)
	default:
		return mcp.NewToolResultErrorf("invalid artifact_type: %s (must be 'layer1', 'layer2', or 'layer3')", artifactType), nil
	}

	if err != nil {
		return mcp.NewToolResultErrorf("failed to build relationships: %v", err), nil
	}

	// Format output
	output, err := marshalOutput(relationships, outputFormat)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal output: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

// ArtifactRelationships represents the relationships for an artifact
type ArtifactRelationships struct {
	ArtifactID   string   `json:"artifact_id" yaml:"artifact_id"`
	ArtifactType string   `json:"artifact_type" yaml:"artifact_type"`
	Title        string   `json:"title,omitempty" yaml:"title,omitempty"`
	ReferencedBy []string `json:"referenced_by,omitempty" yaml:"referenced_by,omitempty"` // Artifacts that reference this one
	References   []string `json:"references,omitempty" yaml:"references,omitempty"`       // Artifacts this one references
	Details      map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

// buildLayer1Relationships finds all Layer 2 controls that reference this Layer 1 guidance
func (g *GemaraAuthoringTools) buildLayer1Relationships(guidanceID string, relationships *ArtifactRelationships) error {
	// Find the Layer 1 guidance
	guidance, exists := g.layer1Guidance[guidanceID]
	if !exists {
		return fmt.Errorf("Layer 1 guidance with ID '%s' not found", guidanceID)
	}

	relationships.Title = guidance.Metadata.Title
	relationships.Details = map[string]interface{}{
		"author":           guidance.Metadata.Author,
		"version":          guidance.Metadata.Version,
		"document_type":    guidance.Metadata.DocumentType,
		"guideline_count":  g.countGuidelines(guidance),
	}

	// Find all Layer 2 controls that reference this guidance
	referencedBy := []string{}
	for catalogID, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for _, control := range family.Controls {
				if g.controlReferencesLayer1(control, guidanceID) {
					referencedBy = append(referencedBy, fmt.Sprintf("%s:%s (Layer 2)", catalogID, control.Id))
				}
			}
		}
	}
	relationships.ReferencedBy = referencedBy

	// Layer 1 doesn't reference other layers
	// Note: Layer 1 guidance documents don't have a Mappings field in the current schema
	relationships.References = []string{}

	return nil
}

// buildLayer2Relationships finds Layer 1 guidance referenced by this control and Layer 3 policies that reference it
func (g *GemaraAuthoringTools) buildLayer2Relationships(controlID string, relationships *ArtifactRelationships) error {
	// Find the Layer 2 control
	var foundControl *layer2.Control
	var catalogID string
	for cid, catalog := range g.layer2Catalogs {
		for _, family := range catalog.ControlFamilies {
			for i := range family.Controls {
				if family.Controls[i].Id == controlID {
					foundControl = &family.Controls[i]
					catalogID = cid
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
		return fmt.Errorf("Layer 2 control with ID '%s' not found", controlID)
	}

	relationships.Title = foundControl.Title
	relationships.Details = map[string]interface{}{
		"catalog_id": catalogID,
		"objective":  foundControl.Objective,
	}

	// Find Layer 1 guidance referenced by this control
	references := []string{}
	if foundControl.GuidelineMappings != nil {
		for _, mapping := range foundControl.GuidelineMappings {
			if mapping.ReferenceId != "" {
				// Check if the referenced guidance exists
				if _, exists := g.layer1Guidance[mapping.ReferenceId]; exists {
					references = append(references, fmt.Sprintf("%s (Layer 1)", mapping.ReferenceId))
				} else {
					references = append(references, fmt.Sprintf("%s (Layer 1 - NOT FOUND)", mapping.ReferenceId))
				}
			}
		}
	}
	relationships.References = references

	// Find Layer 3 policies that reference this control
	referencedBy := []string{}
	for policyID, policy := range g.layer3Policies {
		if g.policyReferencesControl(policy, catalogID, controlID) {
			referencedBy = append(referencedBy, fmt.Sprintf("%s (Layer 3)", policyID))
		}
	}
	relationships.ReferencedBy = referencedBy

	return nil
}

// buildLayer3Relationships finds Layer 1 and Layer 2 artifacts referenced by this policy
func (g *GemaraAuthoringTools) buildLayer3Relationships(policyID string, relationships *ArtifactRelationships) error {
	// Find the Layer 3 policy
	policy, exists := g.layer3Policies[policyID]
	if !exists {
		return fmt.Errorf("Layer 3 policy with ID '%s' not found", policyID)
	}

	relationships.Title = policy.Metadata.Title
	relationships.Details = map[string]interface{}{
		"organization": policy.Metadata.OrganizationID,
		"version":      policy.Metadata.Version,
		"objective":    policy.Metadata.Objective,
	}

	// Find Layer 1 and Layer 2 artifacts referenced by this policy
	references := []string{}
	
	// Layer 1 references
	if policy.GuidanceReferences != nil {
		for _, ref := range policy.GuidanceReferences {
			if ref.ReferenceId != "" {
				if _, exists := g.layer1Guidance[ref.ReferenceId]; exists {
					references = append(references, fmt.Sprintf("%s (Layer 1)", ref.ReferenceId))
				} else {
					references = append(references, fmt.Sprintf("%s (Layer 1 - NOT FOUND)", ref.ReferenceId))
				}
			}
		}
	}

	// Layer 2 references
	// ControlReferences is []Mapping where Mapping.ReferenceId contains the control ID
	// Control IDs may be globally unique or catalog-specific, so we search across all catalogs
	if policy.ControlReferences != nil {
		for _, ref := range policy.ControlReferences {
			if ref.ReferenceId != "" {
				controlID := ref.ReferenceId
				// Search across all catalogs for the control
				found := false
				for catID, catalog := range g.layer2Catalogs {
					for _, family := range catalog.ControlFamilies {
						for _, control := range family.Controls {
							if control.Id == controlID {
								references = append(references, fmt.Sprintf("%s:%s (Layer 2)", catID, controlID))
								found = true
								break
							}
						}
						if found {
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					references = append(references, fmt.Sprintf("%s (Layer 2 - CONTROL NOT FOUND)", controlID))
				}
			}
		}
	}

	relationships.References = references
	// Layer 3 policies don't typically get referenced by other artifacts
	relationships.ReferencedBy = []string{}

	return nil
}

// Helper functions

func (g *GemaraAuthoringTools) countGuidelines(guidance *layer1.GuidanceDocument) int {
	count := 0
	if guidance.Categories != nil {
		for _, category := range guidance.Categories {
			if category.Guidelines != nil {
				count += len(category.Guidelines)
			}
		}
	}
	return count
}

func (g *GemaraAuthoringTools) controlReferencesLayer1(control layer2.Control, guidanceID string) bool {
	if control.GuidelineMappings == nil {
		return false
	}
	for _, mapping := range control.GuidelineMappings {
		if mapping.ReferenceId == guidanceID {
			return true
		}
	}
	return false
}

func (g *GemaraAuthoringTools) policyReferencesControl(policy *layer3.PolicyDocument, catalogID, controlID string) bool {
	if policy.ControlReferences == nil {
		return false
	}
	// ControlReferences is []Mapping where Mapping.ReferenceId contains the control ID
	// We match by control ID (catalogID is provided for context but control IDs should be unique)
	for _, ref := range policy.ControlReferences {
		if ref.ReferenceId == controlID {
			return true
		}
	}
	return false
}
