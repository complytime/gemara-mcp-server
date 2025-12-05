# Testing the Gemara MCP Server

This document provides instructions for testing the Gemara MCP Server.

## Prerequisites

- Go 1.24+ installed
- Python 3.6+ (for the test script)
- Cursor IDE (for MCP integration testing)

## Building the Server

First, build the MCP server binary:

```bash
go build -o gemara-mcp-server ./cmd/gemara-mcp-server
```

Or use the test script with the `--build` flag (see below).

## Testing Methods

### 1. Automated Testing with Python Script

The `test_mcp_server.py` script provides automated testing of the MCP server:

```bash
# Build and test
python3 test_mcp_server.py --build

# Or test with existing binary
python3 test_mcp_server.py --server ./gemara-mcp-server
```

The script will:
1. Start the MCP server
2. Send initialization requests
3. Test listing prompts
4. Test retrieving the Gemara system prompt
5. Display test results

### 2. Manual Testing with stdio

You can manually test the server by sending JSON-RPC messages:

```bash
# Start the server
./gemara-mcp-server

# Then send JSON-RPC messages via stdin, for example:
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./gemara-mcp-server
```

### 3. Testing with Cursor IDE

Cursor has built-in support for MCP servers. To test the server in Cursor:

#### Step 1: Configure Cursor

1. Open Cursor settings
2. Navigate to MCP settings (or create/edit `.cursor/mcp.json` in your workspace)
3. Add the following configuration:

```json
{
  "mcpServers": {
    "gemara-mcp-server": {
      "command": "/absolute/path/to/gemara-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

**Important**: Use the absolute path to the `gemara-mcp-server` binary.

#### Step 2: Restart Cursor

Restart Cursor to load the MCP server configuration.

#### Step 3: Verify Server Connection

1. Open the Cursor command palette (Cmd/Ctrl + Shift + P)
2. Look for MCP-related commands or check the status bar
3. The server should appear as connected

#### Step 4: Test Prompts

Once connected, you can test the prompts:

1. In a chat conversation, the MCP server should provide context about Gemara
2. Try asking questions about:
   - "What is Gemara?"
   - "Explain the 6-layer model"
   - "What are Layer 2 controls?"

The server should provide contextual information from the `gemara-context.md` file.

### 4. Unit Testing (Go)

Run Go unit tests:

```bash
go test ./...
```

To run tests with verbose output:

```bash
go test -v ./...
```

## Expected Behavior

### Successful Server Start

When the server starts correctly:
- It should listen on stdin/stdout for JSON-RPC messages
- No errors should be printed to stderr
- The server should respond to valid JSON-RPC requests

### Initialize Request

The server should respond to an `initialize` request with:
- Server information including name and version
- Capabilities (prompts support)

### List Prompts

The `prompts/list` request should return:
- At least one prompt: `gemara-system-prompt`
- Prompt description: "Provides system-level context about Gemara..."

### Get Prompt

The `prompts/get` request for `gemara-system-prompt` should return:
- Prompt name: "gemara-system-prompt"
- Messages containing the Gemara context markdown content

## Troubleshooting

### Server won't start

- Check that the binary exists and is executable: `ls -l gemara-mcp-server`
- Verify Go version: `go version` (should be 1.24+)
- Check for build errors: `go build ./cmd/gemara-mcp-server`

### No response from server

- Ensure you're sending valid JSON-RPC 2.0 messages
- Check that messages end with a newline (`\n`)
- Verify the server process is running: `ps aux | grep gemara-mcp-server`

### Cursor not connecting

- Verify the absolute path in `.cursor/mcp.json` is correct
- Check Cursor logs for MCP connection errors
- Ensure the binary is executable: `chmod +x gemara-mcp-server`
- Try restarting Cursor completely

### Test script fails

- Ensure Python 3.6+ is installed: `python3 --version`
- Check that the server binary path is correct
- Verify the server builds successfully: `go build ./cmd/gemara-mcp-server`

## Debug Mode

To see more detailed output from the server, you can check stderr. The server uses structured logging with `slog`.

For verbose testing, you can modify the test script to capture and display stderr output.

## Next Steps

- Add more prompts to test
- Implement tools/resource handlers
- Add integration tests for specific use cases
- Test with real Cursor conversations
