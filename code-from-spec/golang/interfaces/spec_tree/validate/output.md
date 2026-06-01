[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@PA5Ahsb3zUIwnpNn8ejRqQpeQNM)

# Package `spectreevalidate`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
```

Package `spectreevalidate` validates the full set of discovered spec tree nodes against structural format rules and returns a list of format errors.

---

## Structs

```go
package spectreevalidate

// SpecTreeValidateInput holds a single discovered node with its parsed
// frontmatter and body, ready for validation.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError describes a single format rule violation found in a node.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

---

## Functions

```go
package spectreevalidate

// SpecTreeValidate validates all entries in the input list against the
// spec tree format rules. It returns a list of FormatError values
// describing every violation found. An empty slice means all nodes are
// valid.
//
// A node is considered to have children when at least one other entry
// in the input list has a logical name that starts with the node's
// logical name followed by "/". A node is a leaf when no entry starts
// with its logical name followed by "/".
func SpecTreeValidate(entries []*SpecTreeValidateInput) []*FormatError
```

---

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
			Frontmatter: &frontmatter.Frontmatter{},
			Node:        &parsenode.Node{},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "interface", Path: "code-from-spec/golang/interfaces/a/b/output.md"},
				},
			},
			Node: &parsenode.Node{},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) == 0 {
		fmt.Println("all nodes are valid")
		return
	}

	for _, e := range errs {
		fmt.Printf("node=%s rule=%s detail=%s\n", e.Node, e.Rule, e.Detail)
	}
}
```
