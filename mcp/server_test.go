package mcp

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewServer(t *testing.T) {
	cfg := &ServerConfig{
		Version:   "test-version",
		Transport: "stdio",
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}

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

func TestHandleGemaraContextResource(t *testing.T) {
	cfg := &ServerConfig{
		Version:   "test-version",
		Transport: "stdio",
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}

	ctx := context.Background()
	request := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "gemara://context/about",
		},
	}

	contents, err := server.handleGemaraContextResource(ctx, request)

	if err != nil {
		t.Fatalf("handleGemaraContextResource returned error: %v", err)
	}

	if contents == nil {
		t.Fatal("handleGemaraContextResource returned nil contents")
	}

	if len(contents) == 0 {
		t.Error("Expected at least one content item in the resource result")
	}

	// Check that gemaraContext is embedded
	if gemaraContext == "" {
		t.Error("gemaraContext should not be empty")
	}

	// Verify first content item contains the context
	if len(contents) > 0 {
		content := contents[0]
		if textContent, ok := content.(*mcp.TextResourceContents); ok {
			if textContent.Text == "" {
				t.Error("Expected non-empty text content")
			}
			if textContent.MIMEType != "text/markdown" {
				t.Errorf("Expected MIME type 'text/markdown', got '%s'", textContent.MIMEType)
			}
		} else {
			t.Error("Expected TextResourceContents type")
		}
	}
}
