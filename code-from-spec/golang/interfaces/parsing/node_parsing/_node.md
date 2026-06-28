---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/parsing/node_parsing/output.md
---

# SPEC/golang/interfaces/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Package

`package parsenode`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"`

## Interface

```go
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
	Public      *NodeSection // nil if absent
	Agent       *NodeSection // nil if absent
	Private     *NodeSection // nil if absent
}

func NodeParse(logicalName string) (*Node, error)
```

`heading` is the normalized form (after `NormalizeText`),
used for comparisons and lookups. `raw_heading` is the
original line as read from the file, preserved for
hashing.

### Errors

- `ErrNotASpecReference`
- `ErrHasQualifier`
- `ErrFileUnreadable`
- `ErrUnexpectedContentBeforeFirstHeading`
- `ErrNodeNameDoesNotMatch`
- `ErrDuplicatePublicSection`
- `ErrDuplicateAgentSection`
- `ErrDuplicatePrivateSection`
- `ErrUnrecognizedSection`
- `ErrDuplicateSubsection`
- Propagated errors from `file` package.

# Agent

Generate an interface specification document listing
the package, import path, struct definitions, and
function signatures.
