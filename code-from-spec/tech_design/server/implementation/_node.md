---
depends_on:
  - ROOT/external/mcp-go-sdk
  - ROOT/tech_design/internal/tools
  - ROOT/tech_design/internal/tools/find_replace
  - ROOT/tech_design/internal/tools/load_chain
  - ROOT/tech_design/internal/tools/write_file
outputs:
  - id: main
    path: cmd/subagent-mcp/main.go
---

# ROOT/tech_design/server/implementation

Generates the main entry point for the MCP server.
