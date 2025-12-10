# Gemara MCP Server

A Model Context Protocol (MCP) server for [Gemara](https://github.com/ossf/gemara) - the GRC Engineering Model for Automated Risk Assessment. This server provides tools and prompts for creating, validating, and managing Gemara artifacts (Layer 1 Guidance, Layer 2 Controls, and Layer 3 Policies).

## Overview

Gemara is a framework for representing cybersecurity guidance, controls, and policies in a structured, machine-readable format. This MCP server enables AI assistants to help users create and manage Gemara artifacts through a standardized interface.

## Quick Start

```bash
# Build the binary
go build ./cmd/gemara-mcp-server

# Run with stdio transport (default)
./gemara-mcp-server
```

For detailed development instructions, containerization, and configuration, see [CURSOR.md](CURSOR.md).

## Features

The server provides tools for:
- **Storage & Validation** - Store and validate Layer 1-3 artifacts
- **Query & Discovery** - List, search, and retrieve artifacts
- **Scoping & Applicability** - Find artifacts applicable to policy scopes
- **File Loading** - Load artifacts from files

See [CURSOR.md](CURSOR.md) for complete tool and prompt documentation.

## License

See [LICENSE](LICENSE) file for details.

## Related Projects

- [Gemara](https://github.com/ossf/gemara) - The core Gemara framework
- [Model Context Protocol](https://modelcontextprotocol.io/) - The MCP specification
