package mcp

import (
	"context"
	_ "embed"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed prompts/gemara-context.md
var gemaraContext string

type ServerConfig struct {
	// Version of the server
	Version string
}

// Server represents the MCP server
type Server struct {
	mcpServer *server.MCPServer
	config    *ServerConfig
}

// NewServer creates a new MCP server
func NewServer(cfg *ServerConfig) *Server {
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

	// Register the Gemara system prompt
	gemaraPrompt := mcp.NewPrompt(
		"gemara-system-prompt",
		mcp.WithPromptDescription("Provides system-level context about Gemara (GRC Engineering Model for Automated Risk Assessment)"),
	)
	s.mcpServer.AddPrompt(gemaraPrompt, s.handleGemaraSystemPrompt)

	return s
}

// Start starts the MCP server
func (s *Server) Start() error {
	return s.ServeStdio()
}

// ServeStdio serves the MCP server via stdio transport
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// handleGemaraSystemPrompt provides system-level context about Gemara to the LLM
func (s *Server) handleGemaraSystemPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return mcp.NewGetPromptResult(
		"Gemara System Context",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(gemaraContext),
			),
		},
	), nil
}
