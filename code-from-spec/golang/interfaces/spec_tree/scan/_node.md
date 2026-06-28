---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/spec_tree/scan/output.md
---

# SPEC/golang/interfaces/spec_tree/scan

Scans the `code-from-spec/` directory and returns all
spec nodes found.

# Public

## Package

`package spectree`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/spectree"`

## Interface

```go
type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}

func SpecTreeScan() ([]*SpecTreeNode, error)
```

Takes no parameters. Scans the `code-from-spec/`
directory relative to the project root. Returns a
list sorted alphabetically by logical name.

### Errors

- `ErrNoNodesFound`: no `_node.md` files found under
  `code-from-spec/`.
- Propagated errors from `listfiles`, `logicalnames`
  packages.

# Agent

Generate an interface specification document listing
the package, import path, struct definition, and
function signatures.
