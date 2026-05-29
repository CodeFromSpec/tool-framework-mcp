[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@ijW1uoflfyGYaULqCvCbyDsXGWY)

# Interface: `spectree`

**Package:** `package spectree`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"`

---

## Structs

```go
// SpecTreeNode represents a single node discovered in the spec tree.
// Each node corresponds to a _node.md file found under code-from-spec/.
type SpecTreeNode struct {
    // LogicalName is the logical name derived from the node's file path.
    LogicalName string

    // FilePath is the CFS path to the _node.md file.
    FilePath *pathutils.PathCfs
}
```

---

## Error Sentinels

No error sentinels are defined by this package. Errors are propagated
from `ListFiles` and `LogicalNameFromPath`.

---

## Functions

```go
// SpecTreeScan scans the code-from-spec/ directory for all _node.md
// files and returns a SpecTreeNode for each one found.
//
// The returned slice is sorted alphabetically by logical name.
//
// Returns an error if:
//   - listing files fails (errors propagated from ListFiles).
//   - deriving a logical name from a path fails (errors propagated
//     from LogicalNameFromPath).
//   - no _node.md files are found under code-from-spec/.
func SpecTreeScan() ([]*SpecTreeNode, error)
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
)

func main() {
    nodes, err := spectree.SpecTreeScan()
    if err != nil {
        log.Fatalf("could not scan spec tree: %v", err)
    }

    for _, node := range nodes {
        fmt.Printf("logical name: %s  path: %s\n", node.LogicalName, node.FilePath.Value)
    }
}
```
