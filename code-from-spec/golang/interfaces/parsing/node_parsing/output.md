[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@I5G2Pp2V_3-14RDQ-PZpCX_zIM0)

# Package `parsenode`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
```

## Structs

```go
package parsenode

type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     []*NodeSection
}
```

## Error Sentinels

```go
package parsenode

import "errors"

var ErrNotARootReference = errors.New("logical name does not start with ROOT/")
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading = errors.New("file body has non-blank content before the first level-1 heading, or has no level-1 heading at all")
var ErrNodeNameDoesNotMatch = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection = errors.New("more than one Agent section exists")
var ErrDuplicateSubsection = errors.New("two level-2 headings within the same section normalize to the same text")
```

## Functions

```go
package parsenode

// NodeParse reads and parses the node file identified by the given logical name.
// It returns a Node containing the name section, optional Public and Agent sections,
// and any additional private sections in file order.
func NodeParse(logical_name string) (*Node, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

func main() {
	node, err := parsenode.NodeParse("ROOT/golang/interfaces/parsing/node_parsing")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Node name:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("Public section found")
		for _, sub := range node.Public.Subsections {
			fmt.Println("  Subsection:", sub.Heading)
		}
	}

	if node.Agent != nil {
		fmt.Println("Agent section found")
	}

	for _, priv := range node.Private {
		fmt.Println("Private section:", priv.Heading)
	}
}
```
