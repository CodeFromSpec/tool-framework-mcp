[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@RKCuJ0xhXXjbDguMwSdyOb29cpY)

# Package `parsenode`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"
```

Package `parsenode` parses a framework node file identified by its logical name into a structured `Node` representation, including named sections and subsections.

---

## Structs

```go
package parsenode

// NodeSubsection represents a level-2 (##) heading block within a section.
// Heading is the normalized form used for comparisons and lookups.
// RawHeading is the original line as read from the file, preserved for hashing.
// Content holds each line of the subsection body as read from the file.
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

// NodeSection represents a level-1 (#) heading block within a node file.
// Heading is the normalized form used for comparisons and lookups.
// RawHeading is the original line as read from the file, preserved for hashing.
// Content holds each line of the section body before the first level-2 heading.
// Subsections holds the ordered list of level-2 blocks within this section.
type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

// Node is the parsed representation of a framework node file.
// NameSection is always present (the first level-1 heading).
// Public and Agent are optional named sections.
// Private holds all other sections in file order.
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

// ErrNotARootReference is returned when the logical name does not
// start with ROOT/.
var ErrNotARootReference = errors.New("logical name does not start with ROOT/")

// ErrHasQualifier is returned when the logical name contains a
// parenthetical qualifier.
var ErrHasQualifier = errors.New("logical name contains a parenthetical qualifier")

// ErrFileUnreadable is returned when the node file cannot be opened
// or read.
var ErrFileUnreadable = errors.New("node file cannot be opened or read")

// ErrUnexpectedContentBeforeFirstHeading is returned when the file
// body has non-blank content before the first level-1 heading, or has
// no level-1 heading at all.
var ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading or no level-1 heading found")

// ErrNodeNameDoesNotMatch is returned when the first heading does not
// match the logical name after normalization.
var ErrNodeNameDoesNotMatch = errors.New("first heading does not match the logical name after normalization")

// ErrDuplicatePublicSection is returned when more than one Public
// section exists in the file.
var ErrDuplicatePublicSection = errors.New("more than one Public section exists")

// ErrDuplicateAgentSection is returned when more than one Agent
// section exists in the file.
var ErrDuplicateAgentSection = errors.New("more than one Agent section exists")

// ErrDuplicateSubsection is returned when two level-2 headings within
// the same section normalize to the same text.
var ErrDuplicateSubsection = errors.New("duplicate subsection heading within a section")
```

---

## Functions

```go
package parsenode

// NodeParse parses the node file identified by logicalName and returns
// a structured Node.
//
// The logical name must begin with ROOT/ and must not contain a
// parenthetical qualifier. The function locates the corresponding file,
// reads it line by line, and constructs the Node from the level-1 and
// level-2 heading structure.
//
// A section that exists in the file but has no content (e.g., # Public
// immediately followed by # Agent) is present with an empty Content
// slice and an empty Subsections slice — it is not absent.
//
// Private sections are returned in the order they appear in the file.
//
// Errors:
//   - ErrNotARootReference: the logical name does not start with ROOT/.
//   - ErrHasQualifier: the logical name contains a parenthetical qualifier.
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrUnexpectedContentBeforeFirstHeading: file body has non-blank
//     content before the first level-1 heading, or has no level-1
//     heading at all. Blank lines before the first heading are not
//     an error.
//   - ErrNodeNameDoesNotMatch: the first heading does not match the
//     logical name after normalization.
//   - ErrDuplicatePublicSection: more than one Public section exists.
//   - ErrDuplicateAgentSection: more than one Agent section exists.
//   - ErrDuplicateSubsection: two level-2 headings within the same
//     section normalize to the same text.
//   - (FileReader.*): propagated from FileOpen.
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
	node, err := parsenode.NodeParse("ROOT/golang/interfaces/parsing/node_parsing")
	if err != nil {
		if errors.Is(err, parsenode.ErrNotARootReference) {
			log.Fatal("logical name must start with ROOT/")
		}
		if errors.Is(err, parsenode.ErrHasQualifier) {
			log.Fatal("logical name must not contain a parenthetical qualifier")
		}
		if errors.Is(err, parsenode.ErrFileUnreadable) {
			log.Fatal("node file could not be opened or read")
		}
		if errors.Is(err, parsenode.ErrUnexpectedContentBeforeFirstHeading) {
			log.Fatal("file has content before the first heading or no heading found")
		}
		if errors.Is(err, parsenode.ErrNodeNameDoesNotMatch) {
			log.Fatal("first heading does not match the logical name")
		}
		if errors.Is(err, parsenode.ErrDuplicatePublicSection) {
			log.Fatal("file has more than one Public section")
		}
		if errors.Is(err, parsenode.ErrDuplicateAgentSection) {
			log.Fatal("file has more than one Agent section")
		}
		if errors.Is(err, parsenode.ErrDuplicateSubsection) {
			log.Fatal("file has a duplicate subsection heading")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Println("node name heading:", node.NameSection.Heading)

	if node.Public != nil {
		fmt.Println("public section heading:", node.Public.Heading)
		for _, sub := range node.Public.Subsections {
			fmt.Println("  subsection:", sub.Heading)
		}
	}

	if node.Agent != nil {
		fmt.Println("agent section heading:", node.Agent.Heading)
	}

	for _, priv := range node.Private {
		fmt.Println("private section:", priv.Heading)
	}
}
```
