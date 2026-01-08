package authoring

import (
	"fmt"
	"log/slog"

	"github.com/complytime/gemara-mcp-server/storage"
	"github.com/complytime/gemara-mcp-server/tools/info"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ossf/gemara"
)

// GemaraAuthoringTools provides tools for creating and validating Gemara artifacts
type GemaraAuthoringTools struct {
	tools     []server.ServerTool
	prompts   []server.ServerPrompt
	resources []server.ServerResource
	// Embedded info tools for validation and schema access
	infoTools *info.GemaraInfoTools
	// Storage interface - can be local or remote
	storage storage.Storage
	// In-memory cache for quick access (populated from storage index)
	layer1Guidance map[string]*gemara.GuidanceDocument
	layer2Catalogs map[string]*gemara.Catalog
	layer3Policies map[string]*gemara.Policy
	// CUE schema cache
	schemaCache map[int]string // layer -> schema content
}

// NewGemaraAuthoringTools creates a new GemaraAuthoringTools instance with default local storage.
func NewGemaraAuthoringTools() (*GemaraAuthoringTools, error) {
	return NewGemaraAuthoringToolsWithStorage(nil)
}

// NewGemaraAuthoringToolsWithStorage creates a new GemaraAuthoringTools instance with the provided storage.
// If storage is nil, it will use the default local file-based storage.
func NewGemaraAuthoringToolsWithStorage(customStorage storage.Storage) (*GemaraAuthoringTools, error) {
	g := &GemaraAuthoringTools{
		layer1Guidance: make(map[string]*gemara.GuidanceDocument),
		layer2Catalogs: make(map[string]*gemara.Catalog),
		layer3Policies: make(map[string]*gemara.Policy),
		schemaCache:    make(map[int]string),
	}

	// Initialize info tools for validation and schema access
	infoTools, err := info.NewGemaraInfoTools()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize info tools: %w", err)
	}
	g.infoTools = infoTools

	// Initialize storage - use custom storage if provided, otherwise use default local storage
	if customStorage != nil {
		g.storage = customStorage
		slog.Info("Using custom storage implementation", "type", "custom")
	} else {
		// Initialize default local storage
		artifactsDir := g.getArtifactsDir()
		slog.Info("Initializing artifact storage", "artifacts_dir", artifactsDir)

		localStorage, err := storage.NewArtifactStorage(artifactsDir)
		if err != nil {
			// Log detailed error for debugging
			slog.Error("Failed to initialize artifact storage",
				"artifacts_dir", artifactsDir,
				"error", err,
			)
			// Return error to fail fast - storage is critical for the server
			return nil, fmt.Errorf("failed to initialize artifact storage at %s: %w", artifactsDir, err)
		}
		g.storage = localStorage
	}

	baseDir := g.storage.GetBaseDir()
	if baseDir != "" {
		slog.Info("Artifact storage initialized successfully", "base_dir", baseDir)
	} else {
		slog.Info("Artifact storage initialized successfully", "type", "remote")
	}

	g.tools = g.registerTools()
	g.prompts = g.registerPrompts()

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
