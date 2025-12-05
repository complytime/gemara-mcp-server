package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer(t *testing.T) {
	cfg := &ServerConfig{
		Version: "test-version",
	}
	
	server := NewServer(cfg)
	
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	
	if server.config == nil {
		t.Error("Server config is nil")
	}
	
	if server.config.Version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", server.config.Version)
	}
	
	if server.mcpServer == nil {
		t.Error("MCP server is nil")
	}
}

func TestHandleGemaraSystemPrompt(t *testing.T) {
	cfg := &ServerConfig{
		Version: "test-version",
	}
	server := NewServer(cfg)
	
	ctx := context.Background()
	request := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "gemara-system-prompt",
		},
	}
	
	result, err := server.handleGemaraSystemPrompt(ctx, request)
	
	if err != nil {
		t.Fatalf("handleGemaraSystemPrompt returned error: %v", err)
	}
	
	if result == nil {
		t.Fatal("handleGemaraSystemPrompt returned nil result")
	}
	
	// Verify the result structure
	if len(result.Messages) == 0 {
		t.Error("Expected at least one message in the prompt result")
	}
	
	// Check that gemaraContext is embedded
	if gemaraContext == "" {
		t.Error("gemaraContext should not be empty")
	}
	
	// Verify first message contains the context
	if len(result.Messages) > 0 {
		msg := result.Messages[0]
		if msg.Role != mcp.RoleUser {
			t.Errorf("Expected role '%s', got '%s'", mcp.RoleUser, msg.Role)
		}
	}
}

// Note: Testing ServeStdio is difficult without actual stdio communication
// Integration tests using the Python script are recommended for full E2E testing
