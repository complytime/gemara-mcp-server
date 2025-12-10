package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerResources registers all resources with the server
func (g *GemaraAuthoringTools) registerResources() []server.ServerResource {
	var resources []server.ServerResource

	// Common schemas
	resources = append(resources, g.newBaseSchemaResource())
	resources = append(resources, g.newMetadataSchemaResource())
	resources = append(resources, g.newMappingSchemaResource())

	// Layer-specific schemas
	resources = append(resources, g.newLayer1SchemaResource())
	resources = append(resources, g.newLayer2SchemaResource())
	resources = append(resources, g.newLayer3SchemaResource())
	resources = append(resources, g.newLayer4SchemaResource())
	resources = append(resources, g.newLayer5SchemaResource())
	resources = append(resources, g.newLayer6SchemaResource())

	return resources
}

// Common Schema Resources

func (g *GemaraAuthoringTools) newBaseSchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/common/base",
			"Base Schema",
			mcp.WithResourceDescription("Common base CUE schema used by all Gemara layers"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleBaseSchemaResource,
	}
}

func (g *GemaraAuthoringTools) newMetadataSchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/common/metadata",
			"Metadata Schema",
			mcp.WithResourceDescription("Common metadata CUE schema used by all Gemara layers"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleMetadataSchemaResource,
	}
}

func (g *GemaraAuthoringTools) newMappingSchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/common/mapping",
			"Mapping Schema",
			mcp.WithResourceDescription("Common mapping CUE schema used by all Gemara layers"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleMappingSchemaResource,
	}
}

// Layer-specific Schema Resources

func (g *GemaraAuthoringTools) newLayer1SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/1",
			"Layer 1 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 1 Guidance documents"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer1SchemaResource,
	}
}

func (g *GemaraAuthoringTools) newLayer2SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/2",
			"Layer 2 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 2 Control Catalogs"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer2SchemaResource,
	}
}

func (g *GemaraAuthoringTools) newLayer3SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/3",
			"Layer 3 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 3 Policy documents"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer3SchemaResource,
	}
}

func (g *GemaraAuthoringTools) newLayer4SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/4",
			"Layer 4 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 4 Evaluation documents"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer4SchemaResource,
	}
}

func (g *GemaraAuthoringTools) newLayer5SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/5",
			"Layer 5 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 5 Enforcement documents"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer5SchemaResource,
	}
}

func (g *GemaraAuthoringTools) newLayer6SchemaResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://schema/layer/6",
			"Layer 6 Schema",
			mcp.WithResourceDescription("CUE schema for Layer 6 Audit documents"),
			mcp.WithMIMEType("text/x-cue"),
		),
		Handler: g.handleLayer6SchemaResource,
	}
}
