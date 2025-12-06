package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
)

// ToolKit represents a collection of MCP functionality.
type ToolKit interface {
	Name() string
	Description() string
	Register(mcpServer *server.MCPServer)
}

// GemaraAuthoringTools provides tools for creating and validating Gemara artifacts
type GemaraAuthoringTools struct {
	tools     []server.ServerTool
	prompts   []server.ServerPrompt
	resources []server.ServerResource
	// Disk-based storage with index
	storage *ArtifactStorage
	// In-memory cache for quick access (populated from storage index)
	layer1Guidance map[string]*layer1.GuidanceDocument
	layer2Catalogs map[string]*layer2.Catalog
	layer3Policies map[string]*layer3.PolicyDocument
	// CUE schema cache
	schemaCache map[int]string // layer -> schema content
}

func NewGemaraAuthoringTools() *GemaraAuthoringTools {
	g := &GemaraAuthoringTools{
		layer1Guidance: make(map[string]*layer1.GuidanceDocument),
		layer2Catalogs: make(map[string]*layer2.Catalog),
		layer3Policies: make(map[string]*layer3.PolicyDocument),
		schemaCache:     make(map[int]string),
	}
	
	// Initialize storage
	artifactsDir := g.getArtifactsDir()
	var err error
	g.storage, err = NewArtifactStorage(artifactsDir)
	if err != nil {
		// If storage initialization fails, log but continue (fallback to in-memory only)
		fmt.Printf("Warning: Failed to initialize artifact storage: %v\n", err)
	}
	
	g.initTools()
	g.initPrompts()
	g.initResources()
	g.LoadArtifactsDir()
	return g
}

func (g *GemaraAuthoringTools) Name() string {
	return "gemara-authoring"
}

func (g *GemaraAuthoringTools) Description() string {
	return "A set of tools related to authoring Gemara artifacts in YAML for Layers 1-4 of the Gemara model."
}

func (g *GemaraAuthoringTools) Register(s *server.MCPServer) {
	s.AddTools(g.tools...)
	s.AddPrompts(g.prompts...)
	s.AddResources(g.resources...)
}

func (g *GemaraAuthoringTools) initTools() {
	g.tools = []server.ServerTool{
		{
			Tool: mcp.NewTool(
				"store_layer1_yaml",
				mcp.WithDescription("Store a Layer 1 Guidance document from raw YAML content. This preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
				mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-1 GuidanceDocument structure. Must include metadata.id and will be validated against the Layer 1 CUE schema."), mcp.Required()),
			),
			Handler: g.handleStoreLayer1YAML,
		},
		{
			Tool: mcp.NewTool(
				"store_layer2_yaml",
				mcp.WithDescription("Store a Layer 2 Control Catalog from raw YAML content. This is the preferred method as it preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
				mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-2 Control Catalog structure. Must include metadata.id and will be validated against the Layer 2 CUE schema."), mcp.Required()),
			),
			Handler: g.handleStoreLayer2YAML,
		},
		{
			Tool: mcp.NewTool(
				"store_layer3_yaml",
				mcp.WithDescription("Store a Layer 3 Policy document from raw YAML content. This is the preferred method as it preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
				mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-3 Policy structure. Must include metadata.id and will be validated against the Layer 3 CUE schema."), mcp.Required()),
			),
			Handler: g.handleStoreLayer3YAML,
		},
		{
			Tool: mcp.NewTool(
				"validate_gemara_yaml",
				mcp.WithDescription("Validate a Gemara YAML document against the appropriate layer schema. Requires the YAML content and layer number (1-4)."),
				mcp.WithString("yaml_content", mcp.Description("The YAML content to validate"), mcp.Required()),
				mcp.WithNumber("layer", mcp.Description("Layer number (1-4) to validate against"), mcp.Required()),
			),
			Handler: g.handleValidateGemaraYAML,
		},
		{
			Tool: mcp.NewTool(
				"get_layer_schema_info",
				mcp.WithDescription("Get information about a Gemara layer schema, including required fields and structure."),
				mcp.WithNumber("layer", mcp.Description("Layer number (1-4) to get schema information for"), mcp.Required()),
			),
			Handler: g.handleGetLayerSchemaInfo,
		},
		{
			Tool: mcp.NewTool(
				"find_applicable_artifacts",
				mcp.WithDescription("Find Layer 1 Guidance documents and Layer 2 Controls that are applicable to a given policy scope. Matches artifacts based on their applicability fields (technology-domains, jurisdictions, industry-sectors for Layer 1; assessment requirement applicability for Layer 2)."),
				mcp.WithArray("boundaries", mcp.Description("Array of boundary/jurisdiction strings (e.g., ['United States', 'European Union'])"), mcp.Items(mcp.WithStringItems())),
				mcp.WithArray("technologies", mcp.Description("Array of technology strings (e.g., ['Payment Processing Systems', 'Cloud Infrastructure'])"), mcp.Items(mcp.WithStringItems())),
				mcp.WithArray("providers", mcp.Description("Array of provider strings (e.g., ['AWS', 'Azure'])"), mcp.Items(mcp.WithStringItems())),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleFindApplicableArtifacts,
		},
		{
			Tool: mcp.NewTool(
				"list_layer1_guidance",
				mcp.WithDescription("List all available Layer 1 Guidance documents. Returns guidance IDs, names, and descriptions that can be referenced when creating Layer 2 controls."),
			),
			Handler: g.handleListLayer1Guidance,
		},
		{
			Tool: mcp.NewTool(
				"get_layer1_guidance",
				mcp.WithDescription("Get detailed information about a specific Layer 1 Guidance document by its guidance_id."),
				mcp.WithString("guidance_id", mcp.Description("The unique identifier of the guidance document"), mcp.Required()),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleGetLayer1Guidance,
		},
		{
			Tool: mcp.NewTool(
				"search_layer1_guidance",
				mcp.WithDescription("Search Layer 1 Guidance documents by name, description, or author. Useful for finding guidance when you know what you're looking for but not the exact ID."),
				mcp.WithString("search_term", mcp.Description("Search term to match against guidance names, descriptions, and authors"), mcp.Required()),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleSearchLayer1Guidance,
		},
		{
			Tool: mcp.NewTool(
				"list_layer2_controls",
				mcp.WithDescription("List available Layer 2 Controls. Can filter by technology or Layer 1 guidance reference. Returns control IDs, names, and descriptions that can be referenced when creating Layer 3 policies."),
				mcp.WithString("technology", mcp.Description("Filter controls by technology (e.g., 'kubernetes', 'docker', 'github')")),
				mcp.WithString("layer1_reference", mcp.Description("Filter controls that reference a specific Layer 1 guidance ID")),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleListLayer2Controls,
		},
		{
			Tool: mcp.NewTool(
				"get_layer2_control",
				mcp.WithDescription("Get detailed information about a specific Layer 2 Control by its control_id."),
				mcp.WithString("control_id", mcp.Description("The unique identifier of the control"), mcp.Required()),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleGetLayer2Control,
		},
		{
			Tool: mcp.NewTool(
				"get_layer2_guideline_mappings",
				mcp.WithDescription("Get all Layer 1 guideline mappings for a Layer 2 control. Shows which Layer 1 guidance documents and specific guidelines are referenced by the control, including strength values and remarks."),
				mcp.WithString("control_id", mcp.Description("The unique identifier of the Layer 2 control"), mcp.Required()),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
				mcp.WithString("include_guidance_details", mcp.Description("Include full Layer 1 guidance document details. Set to 'true' to enable.")),
			),
			Handler: g.handleGetLayer2GuidelineMappings,
		},
		{
			Tool: mcp.NewTool(
				"search_layer2_controls",
				mcp.WithDescription("Search Layer 2 Controls by name or description. Useful for finding controls when you know what you're looking for but not the exact ID."),
				mcp.WithString("search_term", mcp.Description("Search term to match against control names and descriptions"), mcp.Required()),
				mcp.WithString("technology", mcp.Description("Optional: Filter by technology")),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleSearchLayer2Controls,
		},
		{
			Tool: mcp.NewTool(
				"load_layer1_from_file",
				mcp.WithDescription("Load a Layer 1 Guidance document from a YAML or JSON file using the Gemara library. The file will be validated and stored for querying."),
				mcp.WithString("file_path", mcp.Description("Path to the YAML or JSON file (supports file:///path/to/file.yaml or https://example.com/file.yaml)"), mcp.Required()),
			),
			Handler: g.handleLoadLayer1FromFile,
		},
		{
			Tool: mcp.NewTool(
				"load_layer2_from_file",
				mcp.WithDescription("Load a Layer 2 Control Catalog from a YAML or JSON file using the Gemara library. The catalog will be validated and stored for querying."),
				mcp.WithString("file_path", mcp.Description("Path to the YAML or JSON file (supports file:///path/to/file.yaml or https://example.com/file.yaml)"), mcp.Required()),
			),
			Handler: g.handleLoadLayer2FromFile,
		},
		{
			Tool: mcp.NewTool(
				"load_layer3_from_file",
				mcp.WithDescription("Load a Layer 3 Policy document from a YAML or JSON file using the Gemara library. The policy will be validated and stored for querying."),
				mcp.WithString("file_path", mcp.Description("Path to the YAML or JSON file (supports file:///path/to/file.yaml or https://example.com/file.yaml)"), mcp.Required()),
			),
			Handler: g.handleLoadLayer3FromFile,
		},
		{
			Tool: mcp.NewTool(
				"get_artifact_relationships",
				mcp.WithDescription("Visualize cross-layer dependencies for artifacts. Shows which artifacts reference a given artifact and which artifacts it references. Supports Layer 1, Layer 2, and Layer 3 artifacts."),
				mcp.WithString("artifact_id", mcp.Description("The unique identifier of the artifact (guidance_id for Layer 1, control_id for Layer 2, policy_id for Layer 3)"), mcp.Required()),
				mcp.WithString("artifact_type", mcp.Description("The type of artifact: 'layer1', 'layer2', or 'layer3'"), mcp.Required(), mcp.WithStringEnumItems([]string{"layer1", "layer2", "layer3"})),
				mcp.WithString("output_format", mcp.Description("Output format: 'yaml' or 'json'"), mcp.WithStringEnumItems([]string{"yaml", "json"})),
			),
			Handler: g.handleGetArtifactRelationships,
		},
	}
}

// initPrompts initializes MCP prompts for creation tasks
func (g *GemaraAuthoringTools) initPrompts() {
	// Prompts are embedded at compile time via prompts.go
	g.prompts = []server.ServerPrompt{
		{
			Prompt: mcp.NewPrompt(
				"create-layer1-guidance",
				mcp.WithPromptDescription("Guide for creating Layer 1 Guidance documents. Provides YAML structure, examples, and best practices."),
			),
			Handler: func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return mcp.NewGetPromptResult(
					"Creating Layer 1 Guidance Documents",
					[]mcp.PromptMessage{
						mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(createLayer1Prompt)),
					},
				), nil
			},
		},
		{
			Prompt: mcp.NewPrompt(
				"create-layer2-controls",
				mcp.WithPromptDescription("Guide for creating Layer 2 Control Catalogs. Provides YAML structure, examples, and best practices."),
			),
			Handler: func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return mcp.NewGetPromptResult(
					"Creating Layer 2 Control Catalogs",
					[]mcp.PromptMessage{
						mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(createLayer2Prompt)),
					},
				), nil
			},
		},
		{
			Prompt: mcp.NewPrompt(
				"create-layer3-policies",
				mcp.WithPromptDescription("Guide for creating Layer 3 Policy documents. Provides YAML structure, examples, and best practices."),
			),
			Handler: func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return mcp.NewGetPromptResult(
					"Creating Layer 3 Policy Documents",
					[]mcp.PromptMessage{
						mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(createLayer3Prompt)),
					},
				), nil
			},
		},
		{
			Prompt: mcp.NewPrompt(
				"gemara-quick-start",
				mcp.WithPromptDescription("Quick start guide for creating your first Gemara artifacts. Provides step-by-step instructions and common workflows."),
			),
			Handler: func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				return mcp.NewGetPromptResult(
					"Gemara Quick Start Guide",
					[]mcp.PromptMessage{
						mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(quickStartPrompt)),
					},
				), nil
			},
		},
	}
}

// LoadArtifactsDir loads Gemara artifacts from the artifacts directory
func (g *GemaraAuthoringTools) LoadArtifactsDir() {
	artifactsDir := g.getArtifactsDir()

	// Load Layer 1 Guidance artifacts
	layer1Dir := filepath.Join(artifactsDir, "layer1")
	if entries, err := os.ReadDir(layer1Dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml") {
				filePath := filepath.Join(layer1Dir, entry.Name())
				// Convert to absolute path for file:// URI
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					continue
				}
				fileURI := fmt.Sprintf("file://%s", absPath)

				guidance := &layer1.GuidanceDocument{}
				if err := guidance.LoadFile(fileURI); err == nil {
					if guidance.Metadata.Id != "" {
						g.layer1Guidance[guidance.Metadata.Id] = guidance
						// Note: Storage index is already loaded, so we don't need to Add() here
					}
				}
				// Silently skip files that fail to load (they might be invalid)
			}
		}
	}

	// Load Layer 2 Control Catalog artifacts
	layer2Dir := filepath.Join(artifactsDir, "layer2")
	if entries, err := os.ReadDir(layer2Dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml" || filepath.Ext(entry.Name()) == ".json") {
				filePath := filepath.Join(layer2Dir, entry.Name())
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					continue
				}
				fileURI := fmt.Sprintf("file://%s", absPath)

				catalog := &layer2.Catalog{}
				if err := catalog.LoadFile(fileURI); err == nil {
					if catalog.Metadata.Id != "" {
						g.layer2Catalogs[catalog.Metadata.Id] = catalog
					}
				}
				// Silently skip files that fail to load
			}
		}
	}

	// Load Layer 3 Policy artifacts
	layer3Dir := filepath.Join(artifactsDir, "layer3")
	if entries, err := os.ReadDir(layer3Dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".yaml" || filepath.Ext(entry.Name()) == ".yml" || filepath.Ext(entry.Name()) == ".json") {
				filePath := filepath.Join(layer3Dir, entry.Name())
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					continue
				}
				fileURI := fmt.Sprintf("file://%s", absPath)

				policy := &layer3.PolicyDocument{}
				if err := policy.LoadFile(fileURI); err == nil {
					if policy.Metadata.Id != "" {
						g.layer3Policies[policy.Metadata.Id] = policy
					}
				}
				// Silently skip files that fail to load
			}
		}
	}
}
