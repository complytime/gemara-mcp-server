package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer4"
)

// retrieveLayer1Guidance retrieves a Layer 1 Guidance document from cache or storage
func (g *GemaraAuthoringTools) retrieveLayer1Guidance(guidanceID string) (*layer1.GuidanceDocument, error) {
	if guidanceID == "" {
		return nil, fmt.Errorf("guidance_id is required")
	}

	// Check cache first
	if guidance, exists := g.layer1Guidance[guidanceID]; exists {
		return guidance, nil
	}

	// Try storage
	if g.storage != nil {
		retrieved, err := g.storage.Retrieve(1, guidanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve guidance document %s: %w", guidanceID, err)
		}
		if guidance, ok := retrieved.(*layer1.GuidanceDocument); ok {
			// Update cache
			g.layer1Guidance[guidanceID] = guidance
			return guidance, nil
		}
		return nil, fmt.Errorf("retrieved artifact is not a Layer 1 Guidance document")
	}

	return nil, fmt.Errorf("guidance document %s not found", guidanceID)
}

// retrieveLayer2Catalog retrieves a Layer 2 Catalog from cache or storage
func (g *GemaraAuthoringTools) retrieveLayer2Catalog(catalogID string) (*layer2.Catalog, error) {
	if catalogID == "" {
		return nil, fmt.Errorf("catalog_id is required")
	}

	// Check cache first
	if catalog, exists := g.layer2Catalogs[catalogID]; exists {
		return catalog, nil
	}

	// Try storage
	if g.storage != nil {
		retrieved, err := g.storage.Retrieve(2, catalogID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve catalog %s: %w", catalogID, err)
		}
		if catalog, ok := retrieved.(*layer2.Catalog); ok {
			// Update cache
			g.layer2Catalogs[catalogID] = catalog
			return catalog, nil
		}
		return nil, fmt.Errorf("retrieved artifact is not a Layer 2 Catalog")
	}

	return nil, fmt.Errorf("catalog %s not found", catalogID)
}

// retrieveLayer4EvaluationLog retrieves a Layer 4 Evaluation Log from storage
func (g *GemaraAuthoringTools) retrieveLayer4EvaluationLog(evaluationID string) (*layer4.EvaluationLog, error) {
	if evaluationID == "" {
		return nil, fmt.Errorf("evaluation_id is required")
	}

	if g.storage == nil {
		return nil, fmt.Errorf("storage is required for Layer 4 evaluation logs")
	}

	retrieved, err := g.storage.Retrieve(4, evaluationID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve evaluation log %s: %w", evaluationID, err)
	}

	if evaluationLog, ok := retrieved.(*layer4.EvaluationLog); ok {
		return evaluationLog, nil
	}

	return nil, fmt.Errorf("retrieved artifact is not a Layer 4 Evaluation Log")
}

// marshalJSON marshals data to indented JSON
func marshalJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// handleExportLayer1ToOSCAL exports a Layer 1 Guidance document to OSCAL format
func (g *GemaraAuthoringTools) handleExportLayer1ToOSCAL(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	guidanceID := request.GetString("guidance_id", "")
	outputFormat := strings.ToLower(request.GetString("output_format", "profile"))
	guidanceDocHref := request.GetString("guidance_doc_href", "")

	// Validate output format
	if outputFormat != "profile" && outputFormat != "catalog" {
		return mcp.NewToolResultErrorf("output_format must be 'profile' or 'catalog', got '%s'", outputFormat), nil
	}

	// Retrieve guidance document
	guidance, err := g.retrieveLayer1Guidance(guidanceID)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to retrieve guidance: %v", err), nil
	}

	// Convert to OSCAL
	var oscalDoc interface{}
	if outputFormat == "catalog" {
		oscalDoc, err = guidance.ToOSCALCatalog()
		if err != nil {
			return mcp.NewToolResultErrorf("failed to convert to OSCAL catalog: %v", err), nil
		}
	} else {
		// Profile format requires HREF
		if guidanceDocHref == "" {
			guidanceDocHref = fmt.Sprintf("gemara://guidance/%s", guidanceID)
		}
		oscalDoc, err = guidance.ToOSCALProfile(guidanceDocHref)
		if err != nil {
			return mcp.NewToolResultErrorf("failed to convert to OSCAL profile: %v", err), nil
		}
	}

	// Marshal to JSON
	jsonOutput, err := marshalJSON(oscalDoc)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal OSCAL: %v", err), nil
	}

	return mcp.NewToolResultText(jsonOutput), nil
}

// handleExportLayer2ToOSCAL exports a Layer 2 Control Catalog to OSCAL format
func (g *GemaraAuthoringTools) handleExportLayer2ToOSCAL(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	catalogID := request.GetString("catalog_id", "")
	controlHREF := request.GetString("control_href", "")

	// Retrieve catalog
	catalog, err := g.retrieveLayer2Catalog(catalogID)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to retrieve catalog: %v", err), nil
	}

	// Set default HREF if not provided
	if controlHREF == "" {
		controlHREF = fmt.Sprintf("gemara://catalog/%s", catalogID)
	}

	// Convert to OSCAL
	oscalCatalog, err := catalog.ToOSCAL(controlHREF)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to convert to OSCAL: %v", err), nil
	}

	// Marshal to JSON
	jsonOutput, err := marshalJSON(oscalCatalog)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal OSCAL: %v", err), nil
	}

	return mcp.NewToolResultText(jsonOutput), nil
}

// handleExportLayer4ToSARIF exports a Layer 4 Evaluation Log to SARIF format
func (g *GemaraAuthoringTools) handleExportLayer4ToSARIF(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	evaluationID := request.GetString("evaluation_id", "")
	artifactURI := request.GetString("artifact_uri", "")
	catalogID := request.GetString("catalog_id", "")

	// Retrieve evaluation log
	evaluationLog, err := g.retrieveLayer4EvaluationLog(evaluationID)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to retrieve evaluation log: %v", err), nil
	}

	// Set default artifact URI if not provided
	if artifactURI == "" {
		artifactURI = fmt.Sprintf("gemara://evaluation/%s", evaluationID)
	}

	// Retrieve catalog if provided (optional, but improves SARIF output)
	var catalog *layer2.Catalog
	if catalogID != "" {
		cat, err := g.retrieveLayer2Catalog(catalogID)
		if err != nil {
			// Log but don't fail - catalog is optional for SARIF conversion
			// The ToSARIF function can work without it, just with less detail
		} else {
			catalog = cat
		}
	}

	// Convert to SARIF
	sarifBytes, err := evaluationLog.ToSARIF(artifactURI, catalog)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to convert to SARIF: %v", err), nil
	}

	return mcp.NewToolResultText(string(sarifBytes)), nil
}
