---
depends_on:
  - ROOT/golang/dependencies/mcp-go-sdk
  - ROOT/golang/implementation/internal/tools
  - ROOT/golang/implementation/internal/tools/load_chain
  - ROOT/golang/implementation/internal/tools/write_file
  - ROOT/golang/implementation/internal/tools/validate_specs
  - ROOT/golang/implementation/internal/tools/hash_fragment
outputs:
  - id: main
    path: cmd/framework-mcp/main.go
---

# ROOT/golang/implementation/server/code

Generates the main entry point for the MCP server.
