[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@FHDzLohZVJV75klbkXJoRnrgbRw)

# Package `spectreevalidate`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
```

Takes the full set of discovered nodes with their parsed frontmatter and body. Returns a list of format errors (empty if all nodes are valid).

---

## Structs

```go
package spectreevalidate

// SpecTreeValidateInput holds a single discovered node with its parsed
// frontmatter and parsed node structure, as input to SpecTreeValidate.
type SpecTreeValidateInput struct {
	// LogicalName is the logical name of the node (e.g. "ROOT/foo/bar").
	LogicalName string

	// Frontmatter is the parsed frontmatter of the node file.
	Frontmatter *frontmatter.Frontmatter

	// Node is the parsed node structure.
	Node *parsenode.Node
}

// FormatError describes a single validation failure for a node.
type FormatError struct {
	// Node is the logical name of the node that failed validation.
	Node string

	// Rule is the name of the rule that was violated.
	Rule string

	// Detail provides additional context about the violation.
	Detail string
}
```

---

## Functions

```go
package spectreevalidate

// SpecTreeValidate validates the full set of discovered nodes.
//
// A node has children if any other entry in the input list has a logical name
// that starts with the node's logical name followed by "/". For example, given
// entries "ROOT/a" and "ROOT/a/b", "ROOT/a" has children. "ROOT/a/b" is a leaf
// if no entry starts with "ROOT/a/b/".
//
// Returns a list of FormatErrors describing all violations found. Returns an
// empty slice when all nodes are valid.
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
	// Build the input list from discovered nodes (frontmatter and node already parsed).
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT/foo",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{Heading: "foo"},
			},
		},
		{
			LogicalName: "ROOT/foo/bar",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{"ROOT/foo"},
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "interface", Path: "code-from-spec/golang/interfaces/foo/bar/output.md"},
				},
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{Heading: "bar"},
			},
		},
	}

	errs := spectreevalidate.SpecTreeValidate(entries)
	if len(errs) == 0 {
		fmt.Println("All nodes are valid.")
		return
	}

	for _, e := range errs {
		fmt.Printf("node: %s  rule: %s  detail: %s\n", e.Node, e.Rule, e.Detail)
	}
}
```
