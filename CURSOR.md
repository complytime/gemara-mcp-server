# CURSOR.md - Development Guide

This guide provides step-by-step instructions for running, testing, and developing the Gemara MCP Server.

## Prerequisites

- **Go 1.24+** - [Install Go](https://go.dev/doc/install)
- **Make** - Usually pre-installed on Linux/macOS
- **Podman or Docker** (optional) - For containerized builds
- **Cursor IDE** (optional) - For MCP client testing

## Quick Start

### 1. Clone and Navigate

```bash
cd gemara-mcp-server
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build the Binary

```bash
make build
# Binary will be in bin/gemara-mcp-server
```

Or build directly:

```bash
go build -o bin/gemara-mcp-server ./cmd/gemara-mcp-server
```

### 4. Run the Server

**Local Development (Stdio Transport):**
```bash
./bin/gemara-mcp-server
# or with debug logging
./bin/gemara-mcp-server --debug
```

**Remote/Sandboxed Environments (StreamableHTTP via Container):**
See [Container Development](#container-development) section below.

## Development Workflow

### Building

```bash
# Build binary
make build

# Build and install to ~/.local/bin
make install

# Build container image
make container-build
```

### Running Locally

**Stdio Transport:**
```bash
# Basic run
./bin/gemara-mcp-server

# With debug logging
./bin/gemara-mcp-server --debug

# Explicit stdio transport
./bin/gemara-mcp-server --transport stdio --debug
```

**Note:** For remote or sandboxed environments, use StreamableHTTP transport via containers (see [Container Development](#container-development) section).

### Testing

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./tools/...

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run goimports (if installed)
goimports -w .

# Check for issues
go vet ./...

# Build to check for compilation errors
go build ./...
```

## Running in Cursor IDE

### Local Development (Stdio Transport)

For local development, use stdio transport:

1. **Update `.cursor/mcp.json`:**
   ```json
   {
     "mcpServers": {
       "gemara-mcp-server": {
         "command": "/absolute/path/to/bin/gemara-mcp-server",
         "args": ["--debug"]
       }
     }
   }
   ```

2. **Restart Cursor IDE** to pick up the MCP server configuration

3. **Test in Cursor:**
   - Open Cursor chat
   - The MCP server tools should be available
   - Try: "List all Layer 1 guidance documents"

### Remote/Sandboxed Environments (StreamableHTTP)

For remote or sandboxed environments, use StreamableHTTP transport via containers. See [Container Development](#container-development) section for setup instructions.

## Container Development

**Use containers for remote or sandboxed environments that require StreamableHTTP transport.**

### Build Container

```bash
# Build image
make container-build
# or
podman build -t gemara-mcp-server:latest -f Containerfile .
```

### Run Container (StreamableHTTP)

**Default (Writable Artifacts):**
```bash
# Ensure artifacts directory exists
mkdir -p artifacts
chmod 755 artifacts

# Run with StreamableHTTP transport (allows storing new artifacts)
make container-run
# or
podman run --rm --userns=keep-id -p 8080:8080 \
  -v "$(pwd)/artifacts:/app/artifacts:z" \
  --user $(id -u):$(id -g) \
  gemara-mcp-server:latest \
  ./gemara-mcp-server --transport=streamable --port=8080 --debug
```

**Read-Only Mode (Query Only):**
```bash
# Ensure artifacts directory exists
mkdir -p artifacts

# Run with read-only artifacts (cannot store new artifacts, query only)
make container-run-readonly
# or
podman run --rm --userns=keep-id -p 8080:8080 \
  -v "$(pwd)/artifacts:/app/artifacts:z,ro" \
  --user $(id -u):$(id -g) \
  gemara-mcp-server:latest \
  ./gemara-mcp-server --transport=streamable --port=8080 --debug
```

**Note:** 
- The `--userns=keep-id` flag ensures the container user matches your host user ID, preventing permission issues when writing to the mounted artifacts directory.
- The `:z` flag (lowercase) sets a shared SELinux context, which is less restrictive than `:Z` (private context).
- Ensure the `artifacts` directory exists and is writable before running the container.

The server will be accessible at `http://localhost:8080/mcp` for StreamableHTTP connections.

**Note:** In read-only mode, tools that store artifacts (`store_layer1_yaml`, `store_layer2_yaml`, `store_layer3_yaml`) will fail with a storage error.

### Configure Cursor for StreamableHTTP

Update `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "gemara-mcp-server": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Clean Up

```bash
make container-clean
# or
podman rmi gemara-mcp-server:latest
```

## Artifacts Directory

The server automatically looks for artifacts in an `artifacts/` directory:

```
artifacts/
├── layer1/
│   └── *.yaml
├── layer2/
│   └── *.yaml
├── layer3/
│   └── *.yaml
└── layer4/
    └── *.yaml
```

**Create the directory structure:**
```bash
mkdir -p artifacts/{layer1,layer2,layer3,layer4}
```

**Note:** The server will create these directories automatically if they don't exist.

## Common Development Tasks

### Adding a New Tool

1. **Create handler function** in appropriate file (e.g., `tools/layer1.go`):
   ```go
   func (g *GemaraAuthoringTools) handleNewTool(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
       // Implementation
   }
   ```

2. **Add tool definition** in `tools/register_tools.go`:
   ```go
   func (g *GemaraAuthoringTools) newNewTool() server.ServerTool {
       return server.ServerTool{
           Tool: mcp.NewTool(
               "new_tool_name",
               mcp.WithDescription("Tool description"),
               // Parameters...
           ),
           Handler: g.handleNewTool,
       }
   }
   ```

3. **Register tool** in `registerTools()`:
   ```go
   tools = append(tools, g.newNewTool())
   ```

4. **Rebuild and test:**
   ```bash
   make build
   ./bin/gemara-mcp-server --debug
   ```

### Adding a New Prompt

1. **Create prompt file** in `tools/prompts/` (e.g., `new-prompt.md`)

2. **Embed prompt** in `tools/prompts/prompts.go`:
   ```go
   //go:embed new-prompt.md
   var NewPrompt string
   ```

3. **Add prompt definition** in `tools/register_prompts.go`:
   ```go
   func (g *GemaraAuthoringTools) newNewPrompt() server.ServerPrompt {
       return server.ServerPrompt{
           Prompt: mcp.NewPrompt(
               "new-prompt",
               mcp.WithPromptDescription("Prompt description"),
           ),
           Handler: func(_ context.Context, _ mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
               return mcp.NewGetPromptResult(
                   "Prompt Title",
                   []mcp.PromptMessage{
                       mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompts.NewPrompt)),
                   },
               ), nil
           },
       }
   }
   ```

4. **Register prompt** in `registerPrompts()`

### Adding a New Resource

1. **Add resource handler** in `tools/resources.go`:
   ```go
   func (g *GemaraAuthoringTools) handleNewResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
       // Implementation
   }
   ```

2. **Add resource definition** in `tools/register_resources.go`:
   ```go
   func (g *GemaraAuthoringTools) newNewResource() server.ServerResource {
       return server.ServerResource{
           Resource: mcp.NewResource(
               "gemara://resource/path",
               "Resource Name",
               mcp.WithResourceDescription("Description"),
               mcp.WithMIMEType("text/plain"),
           ),
           Handler: g.handleNewResource,
       }
   }
   ```

3. **Register resource** in `registerResources()`

## Debugging

### Enable Debug Logging

```bash
./bin/gemara-mcp-server --debug
```

### Check Logs

Debug logs are written to `stderr`. For stdio transport, logs may be mixed with protocol messages.

For StreamableHTTP, logs appear in the terminal:

```bash
./bin/gemara-mcp-server --transport streamable --port 8080 --debug
```

### Common Issues

**Issue: "Storage not available"**
- **Solution:** Ensure the artifacts directory exists and is writable
- **Check:** `ls -la artifacts/`

**Issue: "Failed to initialize artifact storage"**
- **Solution:** Check directory permissions
- **Fix:** `chmod 755 artifacts/`

**Issue: "Port already in use" (Container)**
- **Solution:** Use a different port or stop the existing container
- **Fix:** Change port mapping: `podman run -p 8081:8080 ...`

**Issue: "CUE validation failed"**
- **Solution:** Check YAML syntax and schema compliance
- **Debug:** Use `validate_gemara_yaml` tool first

**Issue: MCP server not appearing in Cursor**
- **Solution:** 
  1. For stdio: Verify binary path in `.cursor/mcp.json` is absolute
  2. For StreamableHTTP: Verify server is running: `curl http://localhost:8080/mcp`
  3. Check `.cursor/mcp.json` configuration matches your setup
  4. Restart Cursor IDE
  5. Check Cursor logs for MCP connection errors

## Version Information

```bash
./bin/gemara-mcp-server version
```