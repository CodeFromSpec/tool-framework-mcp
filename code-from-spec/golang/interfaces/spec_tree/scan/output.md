[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/scan@vftWbTD-g7JG2wH_nSIfob9hay8)

# Package `spectree`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectree"`

## Package Declaration

```go
package spectree
```

## Struct Definitions

```go
package spectree

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

type SpecTreeNode struct {
	LogicalName string
	FilePath    pathutils.PathCfs
}
```

## Error Sentinels

```go
package spectree

import "errors"

var ErrNoNodesFound = errors.New("no _node.md files found under code-from-spec/")
```

## Function Signatures

```go
package spectree

func SpecTreeScan() ([]*SpecTreeNode, error)
```

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

	for _, n := range nodes {
		fmt.Printf("%s -> %s\n", n.LogicalName, n.FilePath.Value)
	}
}
```
