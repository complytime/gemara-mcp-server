package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initResources initializes MCP resources for CUE schemas
func (g *GemaraAuthoringTools) initResources() {
	g.resources = []server.ServerResource{
		// Common schemas
		{
			Resource: mcp.NewResource(
				"gemara://schema/common/base",
				"Base Schema",
				mcp.WithResourceDescription("Common base CUE schema used by all Gemara layers"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleBaseSchemaResource,
		},
		{
			Resource: mcp.NewResource(
				"gemara://schema/common/metadata",
				"Metadata Schema",
				mcp.WithResourceDescription("Common metadata CUE schema used by all Gemara layers"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleMetadataSchemaResource,
		},
		{
			Resource: mcp.NewResource(
				"gemara://schema/common/mapping",
				"Mapping Schema",
				mcp.WithResourceDescription("Common mapping CUE schema used by all Gemara layers"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleMappingSchemaResource,
		},
		// Layer-specific schemas
		{
			Resource: mcp.NewResource(
				"gemara://schema/layer/1",
				"Layer 1 Schema",
				mcp.WithResourceDescription("CUE schema for Layer 1 Guidance documents"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleLayer1SchemaResource,
		},
		{
			Resource: mcp.NewResource(
				"gemara://schema/layer/2",
				"Layer 2 Schema",
				mcp.WithResourceDescription("CUE schema for Layer 2 Control Catalogs"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleLayer2SchemaResource,
		},
		{
			Resource: mcp.NewResource(
				"gemara://schema/layer/3",
				"Layer 3 Schema",
				mcp.WithResourceDescription("CUE schema for Layer 3 Policy documents"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleLayer3SchemaResource,
		},
		{
			Resource: mcp.NewResource(
				"gemara://schema/layer/4",
				"Layer 4 Schema",
				mcp.WithResourceDescription("CUE schema for Layer 4 Evaluation documents"),
				mcp.WithMIMEType("text/x-cue"),
			),
			Handler: g.handleLayer4SchemaResource,
		},
	}
}

// handleBaseSchemaResource returns the base.cue schema content
func (g *GemaraAuthoringTools) handleBaseSchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCommonCUESchema("base.cue")
	if err != nil {
		return nil, fmt.Errorf("failed to load base schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleMetadataSchemaResource returns the metadata.cue schema content
func (g *GemaraAuthoringTools) handleMetadataSchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCommonCUESchema("metadata.cue")
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleMappingSchemaResource returns the mapping.cue schema content
func (g *GemaraAuthoringTools) handleMappingSchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCommonCUESchema("mapping.cue")
	if err != nil {
		return nil, fmt.Errorf("failed to load mapping schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleLayer1SchemaResource returns the layer-1.cue schema content
func (g *GemaraAuthoringTools) handleLayer1SchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCUESchema(1)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer 1 schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleLayer2SchemaResource returns the layer-2.cue schema content
func (g *GemaraAuthoringTools) handleLayer2SchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCUESchema(2)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer 2 schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleLayer3SchemaResource returns the layer-3.cue schema content
func (g *GemaraAuthoringTools) handleLayer3SchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCUESchema(3)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer 3 schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}

// handleLayer4SchemaResource returns the layer-4.cue schema content
func (g *GemaraAuthoringTools) handleLayer4SchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	schemaContent, err := g.getCUESchema(4)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer 4 schema: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/x-cue",
			Text:     schemaContent,
		},
	}, nil
}
