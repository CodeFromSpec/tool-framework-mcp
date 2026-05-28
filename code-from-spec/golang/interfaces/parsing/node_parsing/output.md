[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@5Ons9P-KENwrh1RKsZl9cZoyf6w)

# Interface: `parsenode`

## Package

```go
package parsenode
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
```

---

## Struct Definitions

```go
// NodeSubsection represents a level-2 heading and its content within
// a section. The heading is stored in normalized form (after NormalizeText).
// Content contains raw markdown text with only leading/trailing blank lines
// trimmed.
type NodeSubsection struct {
	Heading string
	Content string
}

// NodeSection represents a level-1 heading section within a node file.
// The heading is stored in normalized form (after NormalizeText). Content
// contains raw markdown text between the section heading and the first
// subsection (or end of section), with only leading/trailing blank lines
// trimmed. A section that exists in the file but has no content is present
// with an empty Content and an empty Subsections slice.
type NodeSection struct {
	Heading     string
	Content     string
	Subsections []*NodeSubsection
}

// Node represents a parsed _node.md file. Each field corresponds to a
// well-known section of the file, except Private which collects all
// remaining sections in the order they appear.
type Node struct {
	// NameSection is the first level-1 section, whose heading matches
	// the logical name after normalization.
	NameSection *NodeSection

	// Public is the `# Public` section, if present.
	Public *NodeSection

	// Agent is the `# Agent` section, if present.
	Agent *NodeSection

	// Private contains all other level-1 sections, in file order.
	Private []*NodeSection
}
```

---

## Error Sentinels

```go
var (
	// ErrNotRootReference is returned when the logical name does not
	// start with "ROOT/".
	ErrNotRootReference = errors.New("not a ROOT reference")

	// ErrHasQualifier is returned when the logical name contains a
	// parenthetical qualifier.
	ErrHasQualifier = errors.New("has qualifier")

	// ErrFileUnreadable is returned when the node file cannot be opened
	// or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrUnexpectedContentBeforeFirstHeading is returned when the file
	// body has non-blank content before the first level-1 heading, or
	// has no level-1 heading at all.
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

	// ErrNodeNameDoesNotMatch is returned when the first heading does
	// not match the logical name after normalization.
	ErrNodeNameDoesNotMatch = errors.New("node name does not match")

	// ErrDuplicatePublicSection is returned when more than one `# Public`
	// section exists in the file.
	ErrDuplicatePublicSection = errors.New("duplicate public section")

	// ErrDuplicateAgentSection is returned when more than one `# Agent`
	// section exists in the file.
	ErrDuplicateAgentSection = errors.New("duplicate agent section")

	// ErrDuplicateSubsection is returned when two `##` headings within
	// the same section normalize to the same text.
	ErrDuplicateSubsection = errors.New("duplicate subsection")
)
```

---

## Functions

```go
// NodeParse parses the _node.md file associated with the given logical name
// and returns a Node representation of its contents.
//
// The logical name must start with "ROOT/" and must not contain a
// parenthetical qualifier. Path errors from FileOpen are propagated.
//
// Headings are stored in normalized form (after NormalizeText). Content
// fields contain raw markdown text with only leading/trailing blank lines
// trimmed. Private sections preserve their file order.
//
// Possible errors:
//   - ErrNotRootReference
//   - ErrHasQualifier
//   - (path errors from FileOpen)
//   - ErrFileUnreadable
//   - ErrUnexpectedContentBeforeFirstHeading
//   - ErrNodeNameDoesNotMatch
//   - ErrDuplicatePublicSection
//   - ErrDuplicateAgentSection
//   - ErrDuplicateSubsection
func NodeParse(logical_name string) (*Node, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/parsenode"
)

func main() {
	// Parse a node file by its logical name.
	node, err := parsenode.NodeParse("ROOT/golang/interfaces/parsing/node_parsing")
	if err != nil {
		log.Fatal(err)
	}

	// Access the name section heading.
	fmt.Println("Node name:", node.NameSection.Heading)

	// Check if the Public section is present.
	if node.Public != nil {
		fmt.Println("Public content:", node.Public.Content)
		for _, sub := range node.Public.Subsections {
			fmt.Printf("  Subsection: %s\n", sub.Heading)
		}
	}

	// Check if the Agent section is present.
	if node.Agent != nil {
		fmt.Println("Agent content:", node.Agent.Content)
	}

	// Iterate over private sections.
	for _, section := range node.Private {
		fmt.Println("Private section:", section.Heading)
	}
}
```
