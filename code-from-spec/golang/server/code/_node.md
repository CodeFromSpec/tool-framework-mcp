---
depends_on:
  - ROOT/dependencies/mcp-go-sdk
  - ROOT/golang/internal/tools
  - ROOT/golang/internal/tools/load_chain
  - ROOT/golang/internal/tools/write_file
outputs:
  - id: main
    path: cmd/framework-mcp/main.go
---

# ROOT/golang/server/code

Generates the main entry point for the MCP server.
