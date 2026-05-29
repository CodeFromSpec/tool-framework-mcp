[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@7Bf5cLcjbUmB6Ka_xesXo_dvAOU)

# Interface: `logicalnames`

## Package

```go
package logicalnames
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
```

---

## Error Sentinels

```go
var (
	// ErrUnsupportedReference is returned when a logical name does not
	// start with ROOT/.
	ErrUnsupportedReference = errors.New("unsupported reference")

	// ErrInvalidPath is returned when the path is not a _node.md file
	// under code-from-spec/.
	ErrInvalidPath = errors.New("invalid path")

	// ErrNoParent is returned when the logical name is ROOT itself.
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
// Path resolution:
//   - ROOT        → code-from-spec/_node.md
//   - ROOT/x/y    → code-from-spec/x/y/_node.md
//   - ROOT/x/y(z) → code-from-spec/x/y/_node.md
//
// Possible errors:
//   - ErrUnsupportedReference
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
//
// Reverse resolution:
//   - code-from-spec/_node.md     → ROOT
//   - code-from-spec/x/_node.md   → ROOT/x
//   - code-from-spec/x/y/_node.md → ROOT/x/y
//
// Possible errors:
//   - ErrInvalidPath
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts
// ROOT/ references.
//
// Examples:
//   - ROOT/x     → ROOT
//   - ROOT/x/y   → ROOT/x
//   - ROOT/x/y(z)→ ROOT/x
//
// Possible errors:
//   - ErrNoParent
//   - ErrNotARootReference
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns an empty string and false if no qualifier is
// present. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - ROOT/x/y(z)      → "z", true
//   - ARTIFACT/x/y(id) → "id", true
//   - ROOT/x/y         → "", false
func LogicalNameGetQualifier(logical_name string) (string, bool)

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - ROOT/x/y(z)      → ROOT/x/y
//   - ARTIFACT/x/y(id) → ARTIFACT/x/y
//   - ROOT/x/y         → ROOT/x/y
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT, ARTIFACT/
// references, and unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/
// references.
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the
// ARTIFACT/ prefix and any qualifier.
//
// Examples:
//   - ARTIFACT/x/y(id) → ROOT/x/y
//   - ARTIFACT/x/y     → ROOT/x/y
//
// Possible errors:
//   - ErrNotAnArtifactReference
func LogicalNameGetArtifactGenerator(logical_name string) (string, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	// Convert a ROOT/ logical name to a PathCfs.
	cfsPath, err := logicalnames.LogicalNameToPath("ROOT/golang/interfaces/utils/logical_names")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Path:", cfsPath.Value)
	// Output: code-from-spec/golang/interfaces/utils/logical_names/_node.md

	// Derive a logical name from a _node.md path.
	path := &pathutils.PathCfs{Value: "code-from-spec/golang/interfaces/utils/logical_names/_node.md"}
	name, err := logicalnames.LogicalNameFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", name)
	// Output: ROOT/golang/interfaces/utils/logical_names

	// Get the parent of a logical name.
	parent, err := logicalnames.LogicalNameGetParent("ROOT/golang/interfaces/utils/logical_names")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent)
	// Output: ROOT/golang/interfaces/utils

	// Strip a qualifier from a logical name.
	stripped := logicalnames.LogicalNameStripQualifier("ROOT/x/y(z)")
	fmt.Println("Stripped:", stripped)
	// Output: ROOT/x/y

	// Extract a qualifier.
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
	fmt.Println("Qualifier:", qualifier, "Present:", ok)
	// Output: id true

	// Get the generator node for an artifact reference.
	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y(id)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator:", generator)
	// Output: ROOT/x/y
}
```
