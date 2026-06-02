[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@PwYvAUI0iOaeg9vI3OeLilwQJ4Y)

# Package `parsenode`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"`

## Package Declaration

```go
package parsenode
```

## Struct Definitions

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
