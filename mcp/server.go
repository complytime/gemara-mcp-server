package mcp

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/complytime/gemara-mcp-server/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed prompts/gemara-context.md
var gemaraContext string

type ServerConfig struct {
	// Version of the server
	Version string
	// Transport mode selection (stdio/streamable)
	Transport string

	// Host for StreamableHTTP transport
	Host string
	// Port for StreamableHTTP transport
	Port int
	// Logger for HTTP server logging
	Logger *slog.Logger
}

// Server represents the MCP server
type Server struct {
	mcpServer *server.MCPServer
	config    *ServerConfig
}

// NewServer creates a new MCP server
func NewServer(cfg *ServerConfig) (*Server, error) {
	slog.Debug("Creating new MCP server", "version", cfg.Version)

	s := &Server{
		config: cfg,
	}

	// Create MCP server following OpenShift MCP patterns
	mcpServer := server.NewMCPServer(
		"gemara-mcp-server",
		cfg.Version,
		server.WithLogging(),
	)

	s.mcpServer = mcpServer
	slog.Debug("MCP server instance created")

	// Register the Gemara context as a resource so it can be automatically discovered
	gemaraContextResource := mcp.NewResource(
		"gemara://context/about",
		"Gemara Overview",
		mcp.WithResourceDescription("Comprehensive overview and context about Gemara (GRC Engineering Model for Automated Risk Assessment)"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.mcpServer.AddResource(gemaraContextResource, s.handleGemaraContextResource)
	slog.Debug("Gemara context resource registered")

	// Register Gemara Authoring Tools
	slog.Debug("Initializing Gemara authoring tools")
	authoringTools, err := tools.NewGemaraAuthoringTools()
	if err != nil {
		slog.Error("Failed to create authoring tools", "error", err)
		return nil, err
	}
	authoringTools.Register(mcpServer)
	slog.Debug("Gemara authoring tools registered successfully")

	return s, nil
}

// Start starts the MCP server
func (s *Server) Start() error {
	switch s.config.Transport {
	case "stdio":
		return s.ServeStdio()
	case "streamable":
		return s.ServeStreamableHTTP()
	default:
		return fmt.Errorf("unsupported transport mode: %s", s.config.Transport)
	}
}

// ServeStdio serves the MCP server via stdio transport
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) ServeStreamableHTTP() error {
	slog.Info("Starting streamable HTTP server", "host", s.config.Host, "port", s.config.Port)

	var opts []server.StreamableHTTPOption
	if s.config.Logger != nil {
		// Convert slog.Logger to util.Logger interface
		adapter := &slogLoggerAdapter{logger: s.config.Logger}
		opts = append(opts, server.WithLogger(adapter))
	}

	httpServer := server.NewStreamableHTTPServer(s.mcpServer, opts...)
	return httpServer.Start(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
}

// handleGemaraContextResource provides the Gemara context as a resource
func (s *Server) handleGemaraContextResource(_ context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/markdown",
			Text:     gemaraContext,
		},
	}, nil
}

// slogLoggerAdapter adapts slog.Logger to the util.Logger interface
type slogLoggerAdapter struct {
	logger *slog.Logger
}

// Infof implements util.Logger interface
func (a *slogLoggerAdapter) Infof(format string, v ...any) {
	a.logger.Info(fmt.Sprintf(format, v...))
}

// Errorf implements util.Logger interface
func (a *slogLoggerAdapter) Errorf(format string, v ...any) {
	a.logger.Error(fmt.Sprintf(format, v...))
}
