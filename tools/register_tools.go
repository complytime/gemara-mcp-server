package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools registers all tools with the server
func (g *GemaraAuthoringTools) registerTools() []server.ServerTool {
	var tools []server.ServerTool

	// Layer 1 Tools
	tools = append(tools, g.newListLayer1GuidanceTool())
	tools = append(tools, g.newGetLayer1GuidanceTool())
	tools = append(tools, g.newSearchLayer1GuidanceTool())
	tools = append(tools, g.newStoreLayer1YAMLTool())

	// Layer 2 Tools
	tools = append(tools, g.newListLayer2ControlsTool())
	tools = append(tools, g.newGetLayer2ControlTool())
	tools = append(tools, g.newSearchLayer2ControlsTool())
	tools = append(tools, g.newStoreLayer2YAMLTool())
	tools = append(tools, g.newGetLayer2GuidelineMappingsTool())

	// Layer 3 Tools
	tools = append(tools, g.newListLayer3PoliciesTool())
	tools = append(tools, g.newGetLayer3PolicyTool())
	tools = append(tools, g.newSearchLayer3PoliciesTool())
	tools = append(tools, g.newStoreLayer3YAMLTool())

	// Validation and Utility Tools
	tools = append(tools, g.newValidateGemaraYAMLTool())
	tools = append(tools, g.newFindApplicableArtifactsTool())

	// Export Tools
	tools = append(tools, g.newExportLayer1ToOSCALTool())
	tools = append(tools, g.newExportLayer2ToOSCALTool())
	tools = append(tools, g.newExportLayer4ToSARIFTool())

	// Parser Tools
	tools = append(tools, g.newSimpleParseTool())
	tools = append(tools, g.newBenchmarkParserTool())
	tools = append(tools, g.newParseAndValidateTool())

	return tools
}

// Layer 1 Tool Definitions

func (g *GemaraAuthoringTools) newListLayer1GuidanceTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_layer1_guidance",
			mcp.WithDescription("List all available Layer 1 Guidance documents. Returns a summary of all stored guidance documents with their IDs, titles, descriptions, and metadata."),
		),
		Handler: g.handleListLayer1Guidance,
	}
}

func (g *GemaraAuthoringTools) newGetLayer1GuidanceTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_layer1_guidance",
			mcp.WithDescription("Get detailed information about a specific Layer 1 Guidance document by its ID. Returns the full guidance document in YAML or JSON format."),
			mcp.WithString("guidance_id", mcp.Description("The unique identifier of the Layer 1 Guidance document to retrieve."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleGetLayer1Guidance,
	}
}

func (g *GemaraAuthoringTools) newSearchLayer1GuidanceTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"search_layer1_guidance",
			mcp.WithDescription("Search Layer 1 Guidance documents by name, description, or author. Can optionally filter by applicability scope (boundaries, technologies, providers)."),
			mcp.WithString("search_term", mcp.Description("Search term to match against title, description, or author. Required unless scoping filters are provided.")),
			mcp.WithArray("boundaries", mcp.Description("Optional array of boundary/jurisdiction filters to apply.")),
			mcp.WithArray("technologies", mcp.Description("Optional array of technology domain filters to apply.")),
			mcp.WithArray("providers", mcp.Description("Optional array of provider/industry sector filters to apply.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleSearchLayer1Guidance,
	}
}

func (g *GemaraAuthoringTools) newStoreLayer1YAMLTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"store_layer1_yaml",
			mcp.WithDescription("Store a Layer 1 Guidance document from raw YAML content. This preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
			mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-1 GuidanceDocument structure. Must include metadata.id and will be validated against the Layer 1 CUE schema."), mcp.Required()),
		),
		Handler: g.handleStoreLayer1YAML,
	}
}

// Layer 2 Tool Definitions

func (g *GemaraAuthoringTools) newListLayer2ControlsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_layer2_controls",
			mcp.WithDescription("List all available Layer 2 Controls with optional filtering by technology or Layer 1 reference. Returns controls grouped by catalog."),
			mcp.WithString("technology", mcp.Description("Optional technology filter to limit results.")),
			mcp.WithString("layer1_reference", mcp.Description("Optional Layer 1 guidance ID to filter controls that reference it.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleListLayer2Controls,
	}
}

func (g *GemaraAuthoringTools) newGetLayer2ControlTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_layer2_control",
			mcp.WithDescription("Get detailed information about a specific Layer 2 Control by its ID. Returns the full control definition in YAML or JSON format."),
			mcp.WithString("control_id", mcp.Description("The unique identifier of the Layer 2 Control to retrieve."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleGetLayer2Control,
	}
}

func (g *GemaraAuthoringTools) newSearchLayer2ControlsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"search_layer2_controls",
			mcp.WithDescription("Search Layer 2 Controls by name, objective, or ID. Can also filter by Layer 1 guidance reference, technology, or applicability scope."),
			mcp.WithString("search_term", mcp.Description("Search term to match against title, objective, or control ID. Required unless other filters are provided.")),
			mcp.WithString("technology", mcp.Description("Optional technology filter.")),
			mcp.WithString("layer1_reference", mcp.Description("Optional Layer 1 guidance ID to filter controls that reference it.")),
			mcp.WithArray("boundaries", mcp.Description("Optional array of boundary/jurisdiction filters to apply.")),
			mcp.WithArray("technologies", mcp.Description("Optional array of technology domain filters to apply.")),
			mcp.WithArray("providers", mcp.Description("Optional array of provider/industry sector filters to apply.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleSearchLayer2Controls,
	}
}

func (g *GemaraAuthoringTools) newStoreLayer2YAMLTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"store_layer2_yaml",
			mcp.WithDescription("Store a Layer 2 Control Catalog from raw YAML content. This preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
			mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-2 Catalog structure. Must include metadata.id and will be validated against the Layer 2 CUE schema."), mcp.Required()),
		),
		Handler: g.handleStoreLayer2YAML,
	}
}

func (g *GemaraAuthoringTools) newGetLayer2GuidelineMappingsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_layer2_guideline_mappings",
			mcp.WithDescription("Retrieve all Layer 1 guideline mappings for a Layer 2 control. Shows which Layer 1 guidance documents the control references and the specific guideline entries."),
			mcp.WithString("control_id", mcp.Description("The unique identifier of the Layer 2 Control to get mappings for."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
			mcp.WithString("include_guidance_details", mcp.Description("Whether to include full Layer 1 guidance document details. Set to 'true' or '1' to enable.")),
		),
		Handler: g.handleGetLayer2GuidelineMappings,
	}
}

// Layer 3 Tool Definitions

func (g *GemaraAuthoringTools) newListLayer3PoliciesTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"list_layer3_policies",
			mcp.WithDescription("List all available Layer 3 Policy documents. Returns a summary of all stored policy documents with their IDs, titles, objectives, and metadata."),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleListLayer3Policies,
	}
}

func (g *GemaraAuthoringTools) newGetLayer3PolicyTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"get_layer3_policy",
			mcp.WithDescription("Get detailed information about a specific Layer 3 Policy document by its ID. Returns the full policy document in YAML or JSON format."),
			mcp.WithString("policy_id", mcp.Description("The unique identifier of the Layer 3 Policy document to retrieve."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleGetLayer3Policy,
	}
}

func (g *GemaraAuthoringTools) newSearchLayer3PoliciesTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"search_layer3_policies",
			mcp.WithDescription("Search Layer 3 Policy documents by title, objective, or other metadata."),
			mcp.WithString("search_term", mcp.Description("Search term to match against title, objective, or policy ID."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleSearchLayer3Policies,
	}
}

func (g *GemaraAuthoringTools) newStoreLayer3YAMLTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"store_layer3_yaml",
			mcp.WithDescription("Store a Layer 3 Policy document from raw YAML content. This preserves all YAML content without data loss. The YAML is validated with CUE before storing."),
			mcp.WithString("yaml_content", mcp.Description("Raw YAML content containing the complete Layer-3 PolicyDocument structure. Must include metadata.id and will be validated against the Layer 3 CUE schema."), mcp.Required()),
		),
		Handler: g.handleStoreLayer3YAML,
	}
}

// Validation and Utility Tool Definitions

func (g *GemaraAuthoringTools) newValidateGemaraYAMLTool() server.ServerTool {
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

func (g *GemaraAuthoringTools) newFindApplicableArtifactsTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"find_applicable_artifacts",
			mcp.WithDescription("Find Layer 1 and Layer 2 artifacts applicable to a given policy scope. Filters artifacts by boundaries, technologies, and providers."),
			mcp.WithArray("boundaries", mcp.Description("Optional array of boundary/jurisdiction filters.")),
			mcp.WithArray("technologies", mcp.Description("Optional array of technology domain filters.")),
			mcp.WithArray("providers", mcp.Description("Optional array of provider/industry sector filters.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'yaml' (default) or 'json'.")),
		),
		Handler: g.handleFindApplicableArtifacts,
	}
}

// Export Tool Definitions

func (g *GemaraAuthoringTools) newExportLayer1ToOSCALTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"export_layer1_to_oscal",
			mcp.WithDescription("Export a Layer 1 Guidance document to OSCAL format. Returns OSCAL Profile or Catalog JSON."),
			mcp.WithString("guidance_id", mcp.Description("The unique identifier of the Layer 1 Guidance document to export."), mcp.Required()),
			mcp.WithString("output_format", mcp.Description("OSCAL output format: 'profile' (default) or 'catalog'.")),
			mcp.WithString("guidance_doc_href", mcp.Description("Optional HREF for the guidance document (used in profile format).")),
		),
		Handler: g.handleExportLayer1ToOSCAL,
	}
}

func (g *GemaraAuthoringTools) newExportLayer2ToOSCALTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"export_layer2_to_oscal",
			mcp.WithDescription("Export a Layer 2 Control Catalog to OSCAL format. Returns OSCAL Catalog JSON."),
			mcp.WithString("catalog_id", mcp.Description("The unique identifier of the Layer 2 Catalog to export."), mcp.Required()),
			mcp.WithString("control_href", mcp.Description("Optional HREF for the control catalog.")),
		),
		Handler: g.handleExportLayer2ToOSCAL,
	}
}

func (g *GemaraAuthoringTools) newExportLayer4ToSARIFTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"export_layer4_to_sarif",
			mcp.WithDescription("Export a Layer 4 Evaluation Log to SARIF format. Returns SARIF JSON."),
			mcp.WithString("evaluation_id", mcp.Description("The unique identifier of the Layer 4 Evaluation Log to export."), mcp.Required()),
			mcp.WithString("artifact_uri", mcp.Description("Optional URI for the artifact being evaluated.")),
			mcp.WithString("catalog_id", mcp.Description("Optional Layer 2 Catalog ID to include control and requirement details in SARIF output.")),
		),
		Handler: g.handleExportLayer4ToSARIF,
	}
}

// Parser Tool Definitions

func (g *GemaraAuthoringTools) newSimpleParseTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"simple_parse",
			mcp.WithDescription("Parse a PDF file using the simple parser (pdftotext). Extracts text content and structures it into blocks (headings, paragraphs, lists). Returns parsing results with metadata, page count, block statistics, and optional full document content. Optionally validates Layer 1 YAML content if provided."),
			mcp.WithString("file_path", mcp.Description("Path to the PDF file to parse."), mcp.Required()),
			mcp.WithString("temp_dir", mcp.Description("Optional temporary directory for intermediate files. Defaults to system temp directory.")),
			mcp.WithBoolean("keep_temp_files", mcp.Description("Whether to keep temporary files after parsing. Defaults to false.")),
			mcp.WithBoolean("include_full_document", mcp.Description("Whether to include the full parsed document in the response. Defaults to false (returns summary only).")),
			mcp.WithString("layer1_yaml_content", mcp.Description("Optional Layer 1 YAML content to validate after parsing. If provided, validation results will be included in the response.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'json' (default) or 'yaml'.")),
		),
		Handler: g.handleSimpleParse,
	}
}

func (g *GemaraAuthoringTools) newBenchmarkParserTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"benchmark_parser",
			mcp.WithDescription("Benchmark parser performance by running multiple iterations and measuring execution time. Returns statistics including min, max, and average duration, along with success/failure rates."),
			mcp.WithString("file_path", mcp.Description("Path to the PDF file to benchmark."), mcp.Required()),
			mcp.WithNumber("iterations", mcp.Description("Number of iterations to run (1-10). Defaults to 3.")),
			mcp.WithString("parser_type", mcp.Description("Parser type to benchmark: 'simple' (default) or 'docling'.")),
			mcp.WithString("temp_dir", mcp.Description("Optional temporary directory for intermediate files. Defaults to system temp directory.")),
			mcp.WithBoolean("include_details", mcp.Description("Whether to include individual run details in the response. Defaults to false.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'json' (default) or 'yaml'.")),
		),
		Handler: g.handleBenchmarkParser,
	}
}

func (g *GemaraAuthoringTools) newParseAndValidateTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			"parse_and_validate",
			mcp.WithDescription("Run the complete parsing pipeline (parse -> segment -> convert -> validate) on a PDF file. Returns results from each stage including timing information and validation results. Performs both programmatic and CUE validation."),
			mcp.WithString("file_path", mcp.Description("Path to the PDF file to parse and validate."), mcp.Required()),
			mcp.WithString("parser_type", mcp.Description("Parser type: 'simple' (default) or 'docling'.")),
			mcp.WithString("segmenter_type", mcp.Description("Segmenter type: 'generic' (default), 'pci-dss', or 'nist-800-53'.")),
			mcp.WithBoolean("strict", mcp.Description("Enable strict validation mode. Defaults to false.")),
			mcp.WithString("temp_dir", mcp.Description("Optional temporary directory for intermediate files. Defaults to system temp directory.")),
			mcp.WithBoolean("include_layer1_document", mcp.Description("Whether to include the full Layer 1 document in the response. Defaults to false.")),
			mcp.WithString("output_format", mcp.Description("Output format: 'json' (default) or 'yaml'.")),
		),
		Handler: g.handleParseAndValidate,
	}
}
