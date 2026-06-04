[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@VJr6zGilaVQ0UjpbVooaErDJ4to)

# Package `spectreevalidate`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
```

## Structs

```go
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter frontmatter.Frontmatter
	Node        parsenode.Node
}

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

## Functions

```go
package spectreevalidate

// SpecTreeValidate takes the full set of discovered nodes with their parsed
// frontmatter and body. Returns a list of format errors (empty if all nodes
// are valid).
//
// A node has children if any other entry in the input list has a logical name
// that starts with its logical name followed by "/". A node is a leaf if no
// entry starts with its logical name followed by "/".
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError
```

## Usage Example

```go
package main

import (
	"fmt"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
)

func main() {
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT/a",
			Frontmatter: frontmatter.Frontmatter{},
			Node:        parsenode.Node{},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: frontmatter.Frontmatter{Output: "some/output.md"},
			Node:        parsenode.Node{},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) == 0 {
		fmt.Println("All nodes are valid.")
		return
	}

	for _, e := range errs {
		fmt.Printf("Node: %s | Rule: %s | Detail: %s\n", e.Node, e.Rule, e.Detail)
	}
}
```
