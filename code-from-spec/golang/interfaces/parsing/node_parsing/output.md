[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@ODYPWPNZ2yO6mqywoIejqaSnMek)

# Package `parsenode`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"`

## Package Declaration

```go
package parsenode
```

## Struct Definitions

```go
package parsenode

// NodeSubsection represents a level-2 heading and its content within a section.
// Heading is the normalized form; RawHeading is the original line as read from
// the file. Content holds each line exactly as read.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading and its content within a node file.
// Heading is the normalized form; RawHeading is the original line as read from
// the file. Content holds lines before the first level-2 heading, each exactly
// as read. Subsections holds one entry per level-2 heading found.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node is the parsed representation of a _node.md file.
// NameSection is the first (name) section. Public, Agent, and Private hold
// the corresponding sections when present. Private preserves file order.
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

var ErrNotARootReference                    = errors.New("logical name does not start with ROOT/")
var ErrHasQualifier                         = errors.New("logical name contains a parenthetical qualifier")
var ErrFileUnreadable                       = errors.New("file cannot be opened or read")
var ErrUnexpectedContentBeforeFirstHeading  = errors.New("unexpected content before first heading")
var ErrNodeNameDoesNotMatch                 = errors.New("first heading does not match logical name")
var ErrDuplicatePublicSection               = errors.New("more than one Public section")
var ErrDuplicateAgentSection                = errors.New("more than one Agent section")
var ErrDuplicateSubsection                  = errors.New("duplicate subsection heading within section")
```

## Function Signatures

```go
package parsenode

// NodeParse reads and parses the _node.md file for the given logical_name.
// The logical name must start with ROOT/ and must not contain a parenthetical
// qualifier.
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
	node, err := parsenode.NodeParse("ROOT/golang/interfaces/os/file_reader")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("name section heading:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("public section subsections:", len(node.Public.Subsections))
		for _, sub := range node.Public.Subsections {
			fmt.Println("  subsection:", sub.Heading)
		}
	}
}
```
