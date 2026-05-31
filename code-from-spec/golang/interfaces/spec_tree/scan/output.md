[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@8LS_1LLQb-_fb2PpM31_Nlky-bg)

# Package `spectree`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
```

Provides a scan of the spec tree rooted at `code-from-spec/`, returning a sorted list of all nodes found.

---

## Structs

```go
package spectree

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// SpecTreeNode represents a single node discovered in the spec tree.
type SpecTreeNode struct {
	// LogicalName is the logical name of the node derived from its path.
	LogicalName string

	// FilePath is the path to the node's _node.md file in CFS format.
	FilePath *pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package spectree

import "errors"

// ErrNoNodesFound is returned when no _node.md files are found under code-from-spec/.
var ErrNoNodesFound = errors.New("no nodes found")
```

---

## Functions

```go
package spectree

// SpecTreeScan scans the code-from-spec/ directory relative to the project root
// and returns a list of all spec tree nodes found.
//
// Each node corresponds to a _node.md file discovered during the scan.
// The returned list is sorted alphabetically by logical name.
//
// Errors:
//   - ErrNoNodesFound: no _node.md files were found under code-from-spec/.
//   - (ListFiles.*): propagated from ListFiles.
//   - (LogicalNames.*): propagated from LogicalNameFromPath.
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
		log.Fatalf("SpecTreeScan: %v", err)
	}

	for _, node := range nodes {
		fmt.Printf("LogicalName: %s  FilePath: %s\n", node.LogicalName, node.FilePath.Value)
	}
}
```
