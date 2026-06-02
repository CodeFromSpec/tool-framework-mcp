[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@Yu_Hy48GL5dkdyHmDRl_DVrJl8Y)

# Package `spectreevalidate`

```go
package spectreevalidate
```

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"`

## Struct Definitions

```go
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

// SpecTreeValidateInput holds a discovered node with its parsed frontmatter
// and body, ready for validation.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter frontmatter.Frontmatter
	Node        parsenode.Node
}

// FormatError describes a single format violation found during spec tree
// validation.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

## Function Signatures

```go
package spectreevalidate

// SpecTreeValidate validates the full set of discovered nodes with their
// parsed frontmatter and body. Returns a list of format errors found across
// all nodes; the list is empty if all nodes are valid.
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
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) == 0 {
		fmt.Println("all nodes valid")
		return
	}
	for _, e := range errs {
		fmt.Printf("node=%s rule=%s detail=%s\n", e.Node, e.Rule, e.Detail)
	}
}
```
