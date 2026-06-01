[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@IoyET-8CAzhbZfLbHTdO2tvYT3k)

# Package `spectree`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
```

Package `spectree` provides functionality to scan the `code-from-spec/` directory and build a tree of spec nodes with their logical names and file paths.

---

## Structs

```go
package spectree

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// SpecTreeNode represents a single node discovered in the spec tree.
// Each node corresponds to a _node.md file found under code-from-spec/.
type SpecTreeNode struct {
	// LogicalName is the canonical logical name of the node
	// (e.g. "ROOT/functional/logic/os/file_reader").
	LogicalName string

	// FilePath is the CFS-format path to the _node.md file
	// (e.g. "code-from-spec/functional/logic/os/file_reader/_node.md").
	FilePath pathutils.PathCfs
}
```

---

## Error Sentinels

```go
package spectree

import "errors"

// ErrNoNodesFound is returned when no _node.md files are found
// under the code-from-spec/ directory.
var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")
```

---

## Functions

```go
package spectree

// SpecTreeScan scans the code-from-spec/ directory relative to the
// project root and returns all discovered spec nodes sorted
// alphabetically by logical name.
//
// Each node corresponds to a _node.md file. The logical name is
// derived from its path relative to the code-from-spec/ directory.
//
// Errors:
//   - ErrNoNodesFound: no _node.md files found under code-from-spec/.
//   - (ListFiles.*): propagated from the internal file listing operation.
//   - (LogicalNames.*): propagated from LogicalNameFromPath.
func SpecTreeScan() ([]*SpecTreeNode, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"
)

func main() {
	// Scan the spec tree rooted at code-from-spec/.
	nodes, err := spectree.SpecTreeScan()
	if err != nil {
		if errors.Is(err, spectree.ErrNoNodesFound) {
			log.Fatal("no spec nodes found — is the code-from-spec/ directory present?")
		}
		log.Fatalf("scan failed: %v", err)
	}

	// Nodes are sorted alphabetically by logical name.
	fmt.Printf("found %d spec nodes:\n", len(nodes))
	for _, node := range nodes {
		fmt.Printf("  logical_name=%-60s  file_path=%s\n",
			node.LogicalName, node.FilePath.Value)
	}
}
```
