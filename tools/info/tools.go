package info

import (
	"context"

	"github.com/complytime-labs/gemara-mcp-server/tools/prompts"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type GemaraInfoTools struct {
	tools     []server.ServerTool
	prompts   []server.ServerPrompt
	resources []server.ServerResource
	// CUE schema cache - key format: "version:layer" or "version:common:name"
	schemaCache map[string]string
	// Schema version (branch/tag) to use when fetching from GitHub
	schemaVersion string
}

// NewGemaraInfoTools creates a new GemaraInfoTools instance with default schema version "main".
func NewGemaraInfoTools() (*GemaraInfoTools, error) {
	return NewGemaraInfoToolsWithVersion("main")
}

// NewGemaraInfoToolsWithVersion creates a new GemaraInfoTools instance with the specified schema version.
// Version can be a branch name (e.g., "main", "develop") or a tag (e.g., "v1.0.0").
func NewGemaraInfoToolsWithVersion(version string) (*GemaraInfoTools, error) {
	if version == "" {
		version = "main"
	}
	g := &GemaraInfoTools{
		schemaCache:  make(map[string]string),
		schemaVersion: version,
	}

	g.tools = g.registerTools()
	g.prompts = g.registerPrompts()
	g.resources = g.registerResources()

	return g, nil
}

func (g *GemaraInfoTools) Name() string {
	return "gemara-info"
}

func (g *GemaraInfoTools) Description() string {
	return "A set of tools for querying Gemara schema information, validation, and resources."
}

func (g *GemaraInfoTools) Register(s *server.MCPServer) {
	s.AddTools(g.tools...)
	s.AddPrompts(g.prompts...)
	s.AddResources(g.resources...)
}

func (g *GemaraInfoTools) registerTools() []server.ServerTool {
	var tools []server.ServerTool
	tools = append(tools, g.newValidateGemaraYAMLTool())
	tools = append(tools, g.newGetGemaraInfoTool())
	return tools
}

func (g *GemaraInfoTools) newValidateGemaraYAMLTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"validate_gemara_yaml",
			mcp.WithDescription("Validate YAML content against a Gemara layer schema using CUE. Returns a detailed validation report with any errors found."),
			mcp.WithString("yaml_content", mcp.Description("Raw YAML content to validate."), mcp.Required()),
			mcp.WithNumber("layer", mcp.Description("Layer number (1-4) to validate against."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'text' (default), 'json', or 'sarif' (Static Analysis Results Interchange Format).")),
		),
		Handler: g.handleValidateGemaraYAML,
	}
}

func (g *GemaraInfoTools) newGetGemaraInfoTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_gemara_info",
			mcp.WithDescription("Get comprehensive information about Gemara (GRC Engineering Model for Automated Risk Assessment). Use this tool when users ask 'What is Gemara?' or need an overview. Returns overview, architecture, layer model, schema information, and integration details."),
			mcp.WithString("output_format", mcp.Description("Output format: 'text' (default) or 'json'.")),
		),
		Handler: g.handleGetGemaraInfo,
	}
}

// registerResources registers all resources with the server
func (g *GemaraInfoTools) registerResources() []server.ServerResource {
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

	// Documentation resources
	resources = append(resources, g.newLexiconResource())

	return resources
}

// Common Schema Resources

func (g *GemaraInfoTools) newBaseSchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newMetadataSchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newMappingSchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer1SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer2SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer3SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer4SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer5SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLayer6SchemaResource() server.ServerResource {
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

func (g *GemaraInfoTools) newLexiconResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"gemara://lexicon",
			"Gemara Lexicon",
			mcp.WithResourceDescription("Gemara terminology and definitions from the official lexicon"),
			mcp.WithMIMEType("text/html"),
		),
		Handler: g.handleLexiconResource,
	}
}

// registerPrompts registers all prompts with the server
func (g *GemaraInfoTools) registerPrompts() []server.ServerPrompt {
	var prompts []server.ServerPrompt

	prompts = append(prompts, g.newGemaraInfoPrompt())

	return prompts
}

func (g *GemaraInfoTools) newGemaraInfoPrompt() server.ServerPrompt {
	return server.ServerPrompt{
		Prompt: mcp.NewPrompt(
			"gemara-info",
			mcp.WithPromptDescription("Comprehensive information about Gemara (GRC Engineering Model for Automated Risk Assessment). Use this prompt when users ask 'What is Gemara?' or need an overview. Provides overview, architecture, layer model, and integration details."),
		),
		Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return mcp.NewGetPromptResult(
				"Gemara Information",
				[]mcp.PromptMessage{
					mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.GemaraContext)),
				},
			), nil
		},
	}
}
