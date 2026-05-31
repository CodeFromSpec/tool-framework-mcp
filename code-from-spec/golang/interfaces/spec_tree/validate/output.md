[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@7JYnYA5dB4CM-i0zxhrYouaAduI)

# Package `spectreevalidate`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/spectreevalidate"
```

Package `spectreevalidate` validates the full set of discovered spec tree nodes, checking each node's parsed frontmatter and body against format rules. It returns a list of format errors describing any violations found.

---

## Structs

```go
package spectreevalidate

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

// SpecTreeValidateInput holds the data for a single spec tree node
// that is to be validated.
type SpecTreeValidateInput struct {
	// LogicalName is the full logical name of the node (e.g. "ROOT/a/b").
	LogicalName string

	// Frontmatter holds the parsed frontmatter for this node.
	Frontmatter *frontmatter.Frontmatter

	// Node holds the parsed node body for this node.
	Node *parsenode.Node
}

// FormatError describes a single format rule violation found during
// validation of a spec tree node.
type FormatError struct {
	// Node is the logical name of the node that violated the rule.
	Node string

	// Rule is the name or identifier of the rule that was violated.
	Rule string

	// Detail provides a human-readable explanation of the violation.
	Detail string
}
```

---

## Functions

```go
package spectreevalidate

// SpecTreeValidate validates the full set of discovered nodes.
//
// It takes the complete list of nodes with their parsed frontmatter and
// body, and returns a list of FormatError values describing any format
// violations found. The returned slice is empty when all nodes are valid.
//
// A node is considered to have children if any other entry in the input
// list has a logical name that starts with the node's logical name
// followed by "/". For example, given entries "ROOT/a" and "ROOT/a/b",
// "ROOT/a" has children. A node is a leaf if no other entry's logical
// name starts with its own logical name followed by "/".
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
	// Build the input list from previously parsed nodes.
	// In practice, these would be populated by the spec tree scan and parsing steps.
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				External:  []*frontmatter.FrontmatterExternal{},
				Input:     "",
				Outputs:   []*frontmatter.FrontmatterOutput{},
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{
					Heading:     "a",
					RawHeading:  "# a",
					Content:     []string{},
					Subsections: []*parsenode.NodeSubsection{},
				},
				Public:  nil,
				Agent:   nil,
				Private: []*parsenode.NodeSection{},
			},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{
				DependsOn: []string{},
				External:  []*frontmatter.FrontmatterExternal{},
				Input:     "",
				Outputs: []*frontmatter.FrontmatterOutput{
					{ID: "interface", Path: "code-from-spec/golang/interfaces/a/b/output.md"},
				},
			},
			Node: &parsenode.Node{
				NameSection: &parsenode.NodeSection{
					Heading:     "a/b",
					RawHeading:  "# a/b",
					Content:     []string{},
					Subsections: []*parsenode.NodeSubsection{},
				},
				Public:  nil,
				Agent:   nil,
				Private: []*parsenode.NodeSection{},
			},
		},
	}

	// Validate all nodes.
	formatErrors := spectreevalidate.SpecTreeValidate(entries)
	if len(formatErrors) == 0 {
		fmt.Println("all nodes are valid")
		return
	}

	for _, fe := range formatErrors {
		fmt.Printf("node=%s rule=%s detail=%s\n", fe.Node, fe.Rule, fe.Detail)
	}
}
```
