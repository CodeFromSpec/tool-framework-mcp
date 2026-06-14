[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@yFCWkhJuW6ihyyeObtJqOfbfmxI)

# Package `parsenode`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode`

---

## Structs

```go
package parsenode

// NodeSubsection represents a level-2 heading and its content within a section.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading and its content within a node file.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node represents a parsed spec node file.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     *NodeSection
}
```

---

## Error Sentinels

```go
package parsenode

import "errors"

var ErrNotASpecReference                    = errors.New("logical name is not a SPEC/ reference")
var ErrHasQualifier                         = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                       = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading  = errors.New("file has non-blank content before the first level-1 heading, or has no level-1 heading")
var ErrNodeNameDoesNotMatch                 = errors.New("first heading does not match the logical name after normalization")
var ErrDuplicatePublicSection               = errors.New("more than one Public section exists")
var ErrDuplicateAgentSection                = errors.New("more than one Agent section exists")
var ErrDuplicatePrivateSection              = errors.New("more than one Private section exists")
var ErrUnrecognizedSection                  = errors.New("unrecognized level-1 heading")
var ErrDuplicateSubsection                  = errors.New("two level-2 headings within the same section normalize to the same text")
```

---

## Functions

```go
package parsenode

// NodeParse reads and parses the spec node file identified by logical_name.
// The logical name must be a SPEC/ reference and must not contain a parenthetical qualifier.
// Returns the parsed Node or an error describing the failure.
func NodeParse(logicalName string) (*Node, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
)

func main() {
	node, err := parsenode.NodeParse("SPEC/payments/fees")
	if err != nil {
		if errors.Is(err, parsenode.ErrNotASpecReference) {
			log.Fatal("not a spec reference")
		}
		if errors.Is(err, parsenode.ErrFileUnreadable) {
			log.Fatal("file unreadable")
		}
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
}
```
