---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/internal/tools
  - ROOT/golang/internal/tools/load_chain
  - ROOT/golang/internal/tools/write_file
  - ROOT/golang/internal/tools/validate_specs
  - ROOT/golang/internal/tools/hash_fragment
outputs:
  - id: main
    path: cmd/framework-mcp/main.go
---

# ROOT/golang/server/code

Generates the main entry point for the MCP server.
