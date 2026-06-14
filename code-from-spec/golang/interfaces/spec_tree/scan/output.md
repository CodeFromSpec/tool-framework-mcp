[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@yoMUKDzaksJWs6cYS2qBxoOgqpk)

# Package `spectree`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree`

---

## Structs

```go
package spectree

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// SpecTreeNode represents a single node discovered in the spec tree.
type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package spectree

import "errors"

var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")
```

---

## Functions

```go
package spectree

// SpecTreeScan scans the code-from-spec/ directory relative to the project
// root and returns all discovered spec tree nodes sorted alphabetically by
// logical name.
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
		log.Fatal(err)
	}

	for _, node := range nodes {
		fmt.Printf("logical_name=%s  file_path=%s\n", node.LogicalName, node.FilePath.Value)
	}
}
```
