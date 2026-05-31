[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@eJOD9TTiTKO0FgJRlThBLDK6GVo)

# Package `parsenode`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
```

Parses a node file identified by its logical name into a structured `Node` representation, including its sections and subsections.

---

## Structs

```go
package parsenode

// NodeSubsection represents a level-2 heading section (##) within a node file.
// heading is the normalized form (after NormalizeText), used for comparisons and lookups.
// raw_heading is the original line as read from the file (e.g. "## Interface ##"), preserved for hashing.
// Content holds the lines of the subsection as read from the file.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 heading section (#) within a node file.
// heading is the normalized form (after NormalizeText), used for comparisons and lookups.
// raw_heading is the original line as read from the file (e.g. "# Public"), preserved for hashing.
// Content holds the lines before the first level-2 heading.
// Subsections holds the ordered list of level-2 sections within this section.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node represents the parsed structure of a node file.
// NameSection is the mandatory first level-1 section matching the logical name.
// Public is the optional "Public" section.
// Agent is the optional "Agent" section.
// Private is the ordered list of all other level-1 sections.
type Node struct {
	NameSection *NodeSection
	Public      *NodeSection
	Agent       *NodeSection
	Private     []*NodeSection
}
```

---

## Error Sentinels

```go
package parsenode

import "errors"

// ErrNotARootReference is returned when the logical name does not start with ROOT/.
var ErrNotARootReference = errors.New("logical name does not start with ROOT/")

// ErrHasQualifier is returned when the logical name contains a parenthetical qualifier.
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file cannot be opened or read")

// ErrUnexpectedContentBeforeFirstHeading is returned when the file body has
// non-blank content before the first level-1 heading, or has no level-1 heading at all.
// Blank lines before the first heading are not an error.
var ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

// ErrNodeNameDoesNotMatch is returned when the first heading does not match
// the logical name after normalization.
var ErrNodeNameDoesNotMatch = errors.New("node name does not match first heading")

// ErrDuplicatePublicSection is returned when more than one Public section exists.
var ErrDuplicatePublicSection = errors.New("duplicate Public section")

// ErrDuplicateAgentSection is returned when more than one Agent section exists.
var ErrDuplicateAgentSection = errors.New("duplicate Agent section")

// ErrDuplicateSubsection is returned when two level-2 headings within the same
// section normalize to the same text.
var ErrDuplicateSubsection = errors.New("duplicate subsection")
```

---

## Functions

```go
package parsenode

// NodeParse parses the node file identified by logical_name and returns a Node.
//
// The logical name is resolved to a file path (e.g. ROOT/foo/bar -> foo/bar/_node.md),
// read line by line, and parsed into sections and subsections.
//
// Blank lines before the first level-1 heading are tolerated.
// A section that exists in the file but has no content is present with empty
// Content and Subsections — it is not absent.
// Private sections preserve the order they appear in the file.
//
// Errors:
//   - ErrNotARootReference: the logical name does not start with ROOT/.
//   - ErrHasQualifier: the logical name contains a parenthetical qualifier.
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrUnexpectedContentBeforeFirstHeading: non-blank content before the first
//     level-1 heading, or no level-1 heading present.
//   - ErrNodeNameDoesNotMatch: the first heading does not match the logical name
//     after normalization.
//   - ErrDuplicatePublicSection: more than one Public section found.
//   - ErrDuplicateAgentSection: more than one Agent section found.
//   - ErrDuplicateSubsection: two level-2 headings in the same section normalize
//     to the same text.
//   - (FileReader.*): propagated from FileOpen.
func NodeParse(logicalName string) (*Node, error)
```

---

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
		log.Fatalf("NodeParse: %v", err)
	}

	fmt.Println("Node name section heading:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("Public section found, subsections:", len(node.Public.Subsections))
		for _, sub := range node.Public.Subsections {
			fmt.Println("  Subsection:", sub.Heading)
		}
	}

	if node.Agent != nil {
		fmt.Println("Agent section raw heading:", node.Agent.RawHeading)
	}

	fmt.Println("Private sections:", len(node.Private))
	for _, priv := range node.Private {
		fmt.Println("  Private section:", priv.Heading)
	}
}
```
