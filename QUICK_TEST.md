# Quick Testing Guide

## Fastest Way to Test

### 1. Build the server
```bash
go build -o ./bin/gemara-mcp-server ./cmd/gemara-mcp-server
```

This will build and test the server automatically.

### 3. Test with Cursor

1. **Update `.cursor/mcp.json`** with the absolute path to your binary:
   ```json
   {
     "mcpServers": {
       "gemara-mcp-server": {
         "command": "./bin/gemara-mcp-server",
         "args": [],
         "env": {}
       }
     }
   }
   ```

2. **Restart Cursor**

3. **Test in chat**: Ask "What is Gemara?" or "Explain the 6-layer model"

The MCP server should provide context from the Gemara documentation.

## Run Unit Tests

```bash
go test ./mcp -v
```

## Manual Testing

```bash
# Start server
./gemara-mcp-server

# In another terminal, send test message
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./gemara-mcp-server
```

See `TESTING.md` for detailed instructions.
