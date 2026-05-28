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
// NodeSubsection represents a level-2 heading and its content within a
// # Public section. The heading is stored in normalized form (after
// NormalizeText). The content field contains raw markdown text with
// leading and trailing blank lines trimmed.
type NodeSubsection struct {
	Heading string
	Content string
}

// NodeSection represents a level-1 section of a node spec file. The
// heading is stored in normalized form (after NormalizeText). The
// content field contains raw markdown text with leading and trailing
// blank lines trimmed. Subsections are only populated for the # Public
// section; in all other sections, ## headings are treated as content.
type NodeSection struct {
	Heading     string
	Content     string
	Subsections []*NodeSubsection
}

// Node represents a fully parsed node spec file. Name_section is always
// present. Public and Agent are optional and are nil when absent.
// Private contains all other level-1 sections in the order they appear
// in the file.
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
var (
	// ErrNotARootReference is returned when the logical name does not
	// start with "ROOT/".
	ErrNotARootReference = errors.New("not a ROOT reference")

	// ErrHasQualifier is returned when the logical name contains a
	// parenthetical qualifier.
	ErrHasQualifier = errors.New("has qualifier")

	// ErrFileUnreadable is returned when the spec file cannot be opened
	// or read.
	ErrFileUnreadable = errors.New("file unreadable")

	// ErrUnexpectedContentBeforeFirstHeading is returned when the file
	// body has non-blank content before the first level-1 heading, or
	// has no level-1 heading at all.
	ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

	// ErrNodeNameDoesNotMatch is returned when the first heading does not
	// match the logical name after normalization.
	ErrNodeNameDoesNotMatch = errors.New("node name does not match")

	// ErrDuplicatePublicSection is returned when more than one # Public
	// section exists in the file.
	ErrDuplicatePublicSection = errors.New("duplicate public section")

	// ErrDuplicateAgentSection is returned when more than one # Agent
	// section exists in the file.
	ErrDuplicateAgentSection = errors.New("duplicate agent section")

	// ErrDuplicateSubsection is returned when two ## headings within
	// # Public normalize to the same text.
	ErrDuplicateSubsection = errors.New("duplicate subsection")
)
```

---

## Functions

```go
// NodeParse parses the spec file for the given logical name and returns
// a structured Node. The logical name must start with "ROOT/" and must
// not contain a parenthetical qualifier.
//
// The file is located using the logical name via the standard framework
// path resolution. Path errors from FileOpen are propagated directly.
//
// Headings are stored in normalized form (after NormalizeText). Content
// fields contain raw markdown text with leading and trailing blank lines
// trimmed.
//
// A section that exists in the file but has no content is present with
// an empty Content and an empty Subsections slice — it is not nil.
//
// Possible errors:
//   - ErrNotARootReference
//   - ErrHasQualifier
//   - ErrFileUnreadable
//   - ErrUnexpectedContentBeforeFirstHeading
//   - ErrNodeNameDoesNotMatch
//   - ErrDuplicatePublicSection
//   - ErrDuplicateAgentSection
//   - ErrDuplicateSubsection
//   - (path errors propagated from FileOpen)
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
	// Parse a node spec file by its logical name.
	node, err := parsenode.NodeParse("ROOT/golang/interfaces/parsing/node_parsing")
	if err != nil {
		log.Fatal(err)
	}

	// Access the name section (always present).
	fmt.Println("Node name:", node.NameSection.Heading)

	// Check if a Public section exists.
	if node.Public != nil {
		fmt.Println("Public content:", node.Public.Content)

		// Iterate over subsections within # Public.
		for _, sub := range node.Public.Subsections {
			fmt.Printf("  Subsection %q: %s\n", sub.Heading, sub.Content)
		}
	}

	// Check if an Agent section exists.
	if node.Agent != nil {
		fmt.Println("Agent content:", node.Agent.Content)
	}

	// Iterate over private sections in file order.
	for _, sec := range node.Private {
		fmt.Printf("Private section %q\n", sec.Heading)
	}
}
```
