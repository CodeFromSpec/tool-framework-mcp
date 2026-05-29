[//]: # (code-from-spec: ROOT/golang/interfaces/spec_tree/validate@fK1rGDcC84p3FQ3vKL1iw28_NCU)

# Interface: `spectreevalidate`

## Package

```go
package spectreevalidate
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/spectreevalidate"
```

---

## Struct Definitions

```go
// SpecTreeValidateInput represents a single discovered node with its parsed
// frontmatter and body, used as input to SpecTreeValidate.
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

// FormatError represents a single format violation found during validation.
// Node identifies the offending logical name, Rule identifies which
// validation rule was violated, and Detail provides a human-readable
// explanation.
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}
```

---

## Functions

```go
// SpecTreeValidate validates the full set of discovered nodes against the
// spec tree format rules. It returns a list of FormatErrors describing every
// violation found. The returned slice is empty when all nodes are valid.
//
// A node is considered to have children if any other entry in the input list
// has a logical name that starts with the node's logical name followed by
// "/". A node is considered a leaf if no other entry starts with its logical
// name followed by "/".
func SpecTreeValidate(entries []*SpecTreeValidateInput) ([]*FormatError, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/frontmatter"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/spectreevalidate"
)

func main() {
	// Build the input list from previously parsed nodes.
	entries := []*spectreevalidate.SpecTreeValidateInput{
		{
			LogicalName: "ROOT/a",
			Frontmatter: &frontmatter.Frontmatter{},
			Node:        &parsenode.Node{},
		},
		{
			LogicalName: "ROOT/a/b",
			Frontmatter: &frontmatter.Frontmatter{},
			Node:        &parsenode.Node{},
		},
	}

	errs, err := spectreevalidate.SpecTreeValidate(entries)
	if err != nil {
		log.Fatal(err)
	}

	if len(errs) == 0 {
		fmt.Println("All nodes are valid.")
		return
	}

	for _, fe := range errs {
		fmt.Printf("node=%s rule=%s detail=%s\n", fe.Node, fe.Rule, fe.Detail)
	}
}
```
