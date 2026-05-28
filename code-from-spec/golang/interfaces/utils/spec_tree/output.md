[//]: # (code-from-spec: ROOT/golang/interfaces/utils/spec_tree@zIjntt37YES3Yw5vwe2lNWxO0Y4)

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
// SpecTreeNode represents a single spec node discovered on disk.
// It pairs the node's logical name (e.g. "ROOT/golang/interfaces/utils/spec_tree")
// with the CFS-format path to its _node.md file.
type SpecTreeNode struct {
	LogicalName string
	FilePath    *pathutils.PathCfs
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
// project root and returns all discovered spec nodes sorted
// alphabetically by logical name.
//
// Each _node.md file found is converted into a SpecTreeNode containing
// its logical name and its CFS file path.
//
// Possible errors:
//   - ErrNoNodesFound — no _node.md files were found under code-from-spec/.
//   - Errors propagated from ListFiles.
//   - Errors propagated from LogicalNameFromPath.
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
	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		fmt.Printf("logical name: %s, path: %s\n", node.LogicalName, node.FilePath.Value)
	}
}
```
