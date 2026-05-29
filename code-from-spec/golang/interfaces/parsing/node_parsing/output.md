[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/node_parsing@4qCfd_gcEx__c0Fo1x4zDyIfmvU)

# Interface: `parsenode`

**Package:** `package parsenode`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/parsenode"`

---

## Structs

```go
// NodeSubsection represents a level-2 (##) heading section within a Node section.
// Heading is the normalized form of the heading text, used for comparisons and
// lookups. RawHeading is the original heading line as read from the file, preserved
// for hashing. Content holds each line of the subsection body exactly as read.
type NodeSubsection struct {
    Heading    string
    RawHeading string
    Content    []string
}

// NodeSection represents a level-1 (#) heading section within a Node.
// Heading is the normalized form of the heading text, used for comparisons and
// lookups. RawHeading is the original heading line as read from the file, preserved
// for hashing. Content holds each line of the section body before the first ##
// heading, exactly as read. Subsections holds the ordered list of ## subsections.
type NodeSection struct {
    Heading     string
    RawHeading  string
    Content     []string
    Subsections []*NodeSubsection
}

// Node represents a parsed spec node file. NameSection is the first level-1
// heading section, whose heading matches the logical name. Public is the optional
// "# Public" section. Agent is the optional "# Agent" section. Private holds all
// other sections in the order they appear in the file.
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
    // ErrNotARootReference is returned when the logical name does not start with "ROOT/".
    ErrNotARootReference = errors.New("not a ROOT reference")

    // ErrHasQualifier is returned when the logical name contains a parenthetical qualifier.
    ErrHasQualifier = errors.New("has qualifier")

    // ErrFileUnreadable is returned when the file cannot be opened or read.
    ErrFileUnreadable = errors.New("file unreadable")

    // ErrUnexpectedContentBeforeFirstHeading is returned when the file body has
    // non-blank content before the first level-1 heading, or has no level-1 heading at all.
    ErrUnexpectedContentBeforeFirstHeading = errors.New("unexpected content before first heading")

    // ErrNodeNameDoesNotMatch is returned when the first heading does not match
    // the logical name after normalization.
    ErrNodeNameDoesNotMatch = errors.New("node name does not match")

    // ErrDuplicatePublicSection is returned when more than one "# Public" section exists.
    ErrDuplicatePublicSection = errors.New("duplicate public section")

    // ErrDuplicateAgentSection is returned when more than one "# Agent" section exists.
    ErrDuplicateAgentSection = errors.New("duplicate agent section")

    // ErrDuplicateSubsection is returned when two ## headings within the same section
    // normalize to the same text.
    ErrDuplicateSubsection = errors.New("duplicate subsection")
)
```

---

## Functions

```go
// NodeParse parses the spec file for the given logical name and returns a Node.
//
// The logical name must start with "ROOT/" and must not contain a parenthetical
// qualifier. The corresponding file is located via FileOpen. The file is parsed
// into sections and subsections according to level-1 (#) and level-2 (##) headings.
//
// Returns one of the following errors if parsing fails:
//   - ErrNotARootReference: the logical name does not start with "ROOT/".
//   - ErrHasQualifier: the logical name contains a parenthetical qualifier.
//   - path errors propagated from FileOpen.
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrUnexpectedContentBeforeFirstHeading: non-blank content appears before the
//     first level-1 heading, or no level-1 heading exists at all.
//   - ErrNodeNameDoesNotMatch: the first heading does not match the logical name
//     after normalization.
//   - ErrDuplicatePublicSection: more than one "# Public" section exists.
//   - ErrDuplicateAgentSection: more than one "# Agent" section exists.
//   - ErrDuplicateSubsection: two ## headings within the same section normalize
//     to the same text.
func NodeParse(logical_name string) (*Node, error)
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
    // Parse a spec node by its logical name.
    node, err := parsenode.NodeParse("ROOT/golang/interfaces/parsing/node_parsing")
    if err != nil {
        log.Fatalf("could not parse node: %v", err)
    }

    // Access the name section heading.
    fmt.Println("Node name:", node.NameSection.Heading)

    // Check whether a Public section is present.
    if node.Public != nil {
        fmt.Println("Public section content lines:", len(node.Public.Content))
        for _, sub := range node.Public.Subsections {
            fmt.Println("  Subsection:", sub.Heading)
        }
    }

    // Check whether an Agent section is present.
    if node.Agent != nil {
        fmt.Println("Agent section found")
    }

    // Iterate over private sections in file order.
    for _, section := range node.Private {
        fmt.Println("Private section:", section.Heading)
    }
}
```
