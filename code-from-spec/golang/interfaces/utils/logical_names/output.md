[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@kUzswJ1wvfdgamuEmwLtLJm1LRU)

# Interface: `logicalnames`

**Package:** `package logicalnames`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"`

---

## Error Sentinels

```go
var (
    // ErrUnsupportedReference is returned when a logical name does not
    // start with ROOT/.
    ErrUnsupportedReference = errors.New("unsupported reference")

    // ErrInvalidPath is returned when a path is not a _node.md file
    // under code-from-spec/.
    ErrInvalidPath = errors.New("invalid path")

    // ErrNoParent is returned when the logical name is ROOT itself and
    // has no parent.
    ErrNoParent = errors.New("no parent")

    // ErrNotARootReference is returned when the logical name does not
    // start with ROOT/.
    ErrNotARootReference = errors.New("not a ROOT reference")

    // ErrNotAnArtifactReference is returned when the logical name does
    // not start with ARTIFACT/.
    ErrNotAnArtifactReference = errors.New("not an artifact reference")
)
```

---

## Functions

```go
// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of
// the corresponding _node.md file. Strips any qualifier before
// resolving. Only accepts ROOT/ references.
//
// Examples:
//   - "ROOT"        → "code-from-spec/_node.md"
//   - "ROOT/x/y"   → "code-from-spec/x/y/_node.md"
//   - "ROOT/x/y(z)"→ "code-from-spec/x/y/_node.md"
//
// Returns ErrUnsupportedReference if the logical name does not start
// with ROOT/.
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
//
// Examples:
//   - "code-from-spec/_node.md"     → "ROOT"
//   - "code-from-spec/x/_node.md"   → "ROOT/x"
//   - "code-from-spec/x/y/_node.md" → "ROOT/x/y"
//
// Returns ErrInvalidPath if the path is not a _node.md file under
// code-from-spec/.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts
// ROOT/ references.
//
// Examples:
//   - "ROOT/x"     → "ROOT"
//   - "ROOT/x/y"   → "ROOT/x"
//   - "ROOT/x/y(z)"→ "ROOT/x"
//
// Returns ErrNoParent if the logical name is ROOT itself.
// Returns ErrNotARootReference if the logical name does not start
// with ROOT/.
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns the empty string and false if no qualifier is
// present. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"     → ("z", true)
//   - "ARTIFACT/x/y(id)"→ ("id", true)
//   - "ROOT/x/y"        → ("", false)
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"      → "ROOT/x/y"
//   - "ARTIFACT/x/y(id)" → "ARTIFACT/x/y"
//   - "ROOT/x/y"         → "ROOT/x/y"
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT,
// ARTIFACT/ references, and unrecognized prefixes.
//
// Examples:
//   - "ROOT/x"       → true
//   - "ROOT/x/y"     → true
//   - "ROOT"         → false
//   - "ARTIFACT/x/y" → false
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/
// references.
//
// Examples:
//   - "ROOT/x/y(z)"     → true
//   - "ARTIFACT/x/y(id)"→ true
//   - "ROOT/x/y"        → false
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
//
// Examples:
//   - "ARTIFACT/x/y" → true
//   - "ROOT/x/y"     → false
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the
// ARTIFACT/ prefix and any qualifier.
//
// Examples:
//   - "ARTIFACT/x/y(id)" → "ROOT/x/y"
//   - "ARTIFACT/x/y"     → "ROOT/x/y"
//
// Returns ErrNotAnArtifactReference if the logical name does not start
// with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logical_name string) (string, error)
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
    "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
    // Convert a ROOT/ logical name to its _node.md PathCfs.
    cfsPath, err := logicalnames.LogicalNameToPath("ROOT/golang/interfaces/utils/logical_names")
    if err != nil {
        log.Fatalf("could not convert logical name to path: %v", err)
    }
    fmt.Println("Node path:", cfsPath.Value)
    // Output: code-from-spec/golang/interfaces/utils/logical_names/_node.md

    // Derive a logical name from a _node.md path.
    logicalName, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{
        Value: "code-from-spec/golang/interfaces/utils/logical_names/_node.md",
    })
    if err != nil {
        log.Fatalf("could not derive logical name from path: %v", err)
    }
    fmt.Println("Logical name:", logicalName)
    // Output: ROOT/golang/interfaces/utils/logical_names

    // Get the parent of a logical name.
    parent, err := logicalnames.LogicalNameGetParent("ROOT/golang/interfaces/utils/logical_names")
    if err != nil {
        log.Fatalf("could not get parent: %v", err)
    }
    fmt.Println("Parent:", parent)
    // Output: ROOT/golang/interfaces/utils

    // Strip a qualifier from a logical name.
    stripped := logicalnames.LogicalNameStripQualifier("ROOT/x/y(z)")
    fmt.Println("Stripped:", stripped)
    // Output: ROOT/x/y

    // Extract the qualifier from a logical name.
    qualifier, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
    if ok {
        fmt.Println("Qualifier:", qualifier)
        // Output: id
    }

    // Resolve an ARTIFACT/ reference to its generating node's logical name.
    generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y(id)")
    if err != nil {
        log.Fatalf("could not get artifact generator: %v", err)
    }
    fmt.Println("Generator:", generator)
    // Output: ROOT/x/y
}
```
