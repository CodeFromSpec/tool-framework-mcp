[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@zIjntt37YES3Yw5vwe2lNWxO0Y4)

# Interface: `spectree`

## Package

```go
package spectree
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/spectree"
```

---

## Struct Definitions

```go
// SpecTreeNode represents a single node discovered in the spec tree.
// It pairs the node's logical name with the CFS path to its _node.md file.
type SpecTreeNode struct {
	// LogicalName is the framework-level identifier for this node,
	// derived from the path of its _node.md file.
	LogicalName string

	// FilePath is the CFS-format path to the _node.md file for this node,
	// relative to the project root.
	FilePath *pathutils.PathCfs
}
```

---

## Error Sentinels

```go
var (
	// ErrNoNodesFound is returned when no _node.md files are found
	// under the code-from-spec/ directory.
	ErrNoNodesFound = errors.New("no nodes found")
)
```

---

## Functions

```go
// SpecTreeScan scans the code-from-spec/ directory relative to the
// project root and returns all discovered spec tree nodes, sorted
// alphabetically by logical name.
//
// Each node corresponds to a _node.md file found in the tree. The
// logical name is derived from the file's path, and the file path is
// stored as a PathCfs.
//
// Possible errors:
//   - ErrNoNodesFound — no _node.md files were found under code-from-spec/
//   - errors propagated from ListFiles
//   - errors propagated from LogicalNameFromPath
func SpecTreeScan() ([]*SpecTreeNode, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/spectree"
)

func main() {
	// Scan the spec tree for all nodes.
	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over nodes sorted alphabetically by logical name.
	for _, node := range nodes {
		fmt.Printf("LogicalName: %s, FilePath: %s\n", node.LogicalName, node.FilePath.Value)
	}
}
```
