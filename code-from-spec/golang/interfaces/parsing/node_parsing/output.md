[//]: # (code-from-spec: SPEC/golang/interfaces/parsing/node_parsing@ydK16VHic1_MNy1I9UeVeNx-g-Y)

# Package `parsenode`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode`

## Types

```go
package parsenode

// NodeSubsection represents a level-2 heading block within a section.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading block within a node file.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node holds the parsed structure of a node file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}
```

## Error Sentinels

```go
package parsenode

import "errors"

var ErrNotASpecReference                    = errors.New("logical name is not a SPEC/ reference")
var ErrHasQualifier                         = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                       = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading  = errors.New("file body has non-blank content before the first level-1 heading, or has no level-1 heading at all")
var ErrNodeNameDoesNotMatch                 = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection               = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection                = errors.New("more than one Agent section exists")
var ErrDuplicatePrivateSection              = errors.New("more than one Private section exists")
var ErrUnrecognizedSection                  = errors.New("unrecognized level-1 heading")
var ErrDuplicateSubsection                  = errors.New("two level-2 headings within the same section normalize to the same text")
```

## Functions

```go
package parsenode

// NodeParse reads and parses the node file for the given logical name.
// The logical name must be a SPEC/ reference and must not contain a parenthetical qualifier.
// Returns a fully populated Node on success, or an error describing the first violation found.
func NodeParse(logicalName string) (*Node, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"
)

func main() {
	node, err := parsenode.NodeParse("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Node name section heading:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("Public section has", len(node.Public.Subsections), "subsection(s)")
		for _, sub := range node.Public.Subsections {
			fmt.Println("  Subsection:", sub.Heading)
		}
	}

	if node.Agent != nil {
		fmt.Println("Agent section content lines:", len(node.Agent.Content))
	}

	if node.Private != nil {
		fmt.Println("Private section content lines:", len(node.Private.Content))
	}
}
```
