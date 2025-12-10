package tools

import (
	"fmt"
	"log/slog"

	"github.com/complytime/gemara-mcp-server/storage"
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
	storage *storage.ArtifactStorage
	// In-memory cache for quick access (populated from storage index)
	layer1Guidance map[string]*layer1.GuidanceDocument
	layer2Catalogs map[string]*layer2.Catalog
	layer3Policies map[string]*layer3.PolicyDocument
	// CUE schema cache
	schemaCache map[int]string // layer -> schema content
}

func NewGemaraAuthoringTools() (*GemaraAuthoringTools, error) {
	g := &GemaraAuthoringTools{
		layer1Guidance: make(map[string]*layer1.GuidanceDocument),
		layer2Catalogs: make(map[string]*layer2.Catalog),
		layer3Policies: make(map[string]*layer3.PolicyDocument),
		schemaCache:    make(map[int]string),
	}

	// Initialize storage
	artifactsDir := g.getArtifactsDir()
	slog.Info("Initializing artifact storage", "artifacts_dir", artifactsDir)
	
	var err error
	g.storage, err = storage.NewArtifactStorage(artifactsDir)
	if err != nil {
		// Log detailed error for debugging
		slog.Error("Failed to initialize artifact storage",
			"artifacts_dir", artifactsDir,
			"error", err,
		)
		// Return error to fail fast - storage is critical for the server
		return nil, fmt.Errorf("failed to initialize artifact storage at %s: %w", artifactsDir, err)
	}

	slog.Info("Artifact storage initialized successfully", "base_dir", g.storage.GetBaseDir())

	g.tools = g.registerTools()
	g.prompts = g.registerPrompts()
	g.resources = g.registerResources()

	// Load artifacts - this may fail if directory doesn't exist, but that's OK
	// LoadArtifactsDir doesn't return an error, it handles failures internally
	g.LoadArtifactsDir()

	return g, nil
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

