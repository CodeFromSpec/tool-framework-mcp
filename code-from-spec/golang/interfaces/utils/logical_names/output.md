[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@zoMfJQHVhwD14N6oZiWbS1d2DH8)

# Package `logicalnames`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
```

Provides utilities for working with logical names in the Code from Spec framework. Logical names use two prefixes — `ROOT/` for spec nodes and `ARTIFACT/` for generated artifacts — and may carry an optional parenthetical qualifier.

---

## Error Sentinels

```go
package logicalnames

import "errors"

// ErrUnsupportedReference is returned when a logical name is not a ROOT/ reference.
var ErrUnsupportedReference = errors.New("unsupported reference: not a ROOT/ reference")

// ErrInvalidPath is returned when a path is not a _node.md file under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent")

// ErrNotARootReference is returned when the logical name is not a ROOT/ reference.
var ErrNotARootReference = errors.New("not a ROOT/ reference")

// ErrNotAnArtifactReference is returned when the logical name does not start with ARTIFACT/.
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")
```

---

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references (including ROOT itself).
//
// Errors:
//   - ErrUnsupportedReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
//
// Errors:
//   - ErrInvalidPath: the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
// Only accepts ROOT/ references (including ROOT itself, which returns NoParent).
//
// Errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameGetParent(logicalName string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical
// name. Returns ("", false) if no qualifier is present. Works with both ROOT/
// and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"      → ("z", true)
//   - "ROOT/x/y"         → ("", false)
//   - "ARTIFACT/x/y(id)" → ("id", true)
func LogicalNameGetQualifier(logicalName string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"       → "ROOT/x/y"
//   - "ARTIFACT/x/y(id)"  → "ARTIFACT/x/y"
//   - "ROOT/x/y"          → "ROOT/x/y"
func LogicalNameStripQualifier(logicalName string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself. Returns false for ROOT, ARTIFACT/ references, and
// unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameHasQualifier(logicalName string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and any
// qualifier.
//
// Examples:
//   - "ARTIFACT/x/y(id)" → "ROOT/x/y"
//   - "ARTIFACT/x/y"     → "ROOT/x/y"
//
// Errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error)
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
	cfs, err := logicalnames.LogicalNameToPath("ROOT/functional/logic")
	if err != nil {
		log.Fatalf("LogicalNameToPath: %v", err)
	}
	fmt.Println("PathCfs:", cfs.Value)
	// Output: code-from-spec/functional/logic/_node.md

	// Convert a _node.md PathCfs back to its logical name.
	nodePath := &pathutils.PathCfs{Value: "code-from-spec/functional/logic/_node.md"}
	name, err := logicalnames.LogicalNameFromPath(nodePath)
	if err != nil {
		log.Fatalf("LogicalNameFromPath: %v", err)
	}
	fmt.Println("Logical name:", name)
	// Output: ROOT/functional/logic

	// Navigate to the parent node.
	parent, err := logicalnames.LogicalNameGetParent("ROOT/functional/logic")
	if err != nil {
		log.Fatalf("LogicalNameGetParent: %v", err)
	}
	fmt.Println("Parent:", parent)
	// Output: ROOT/functional

	// Extract and strip a qualifier.
	qualified := "ROOT/x/y(z)"
	if qualifier, ok := logicalnames.LogicalNameGetQualifier(qualified); ok {
		fmt.Println("Qualifier:", qualifier)
		// Output: z
	}
	fmt.Println("Stripped:", logicalnames.LogicalNameStripQualifier(qualified))
	// Output: ROOT/x/y

	// Resolve an ARTIFACT/ reference to its generating node's logical name.
	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y(id)")
	if err != nil {
		log.Fatalf("LogicalNameGetArtifactGenerator: %v", err)
	}
	fmt.Println("Generator:", generator)
	// Output: ROOT/x/y

	// Predicate helpers.
	fmt.Println(logicalnames.LogicalNameHasParent("ROOT/x/y"))  // true
	fmt.Println(logicalnames.LogicalNameHasParent("ROOT"))       // false
	fmt.Println(logicalnames.LogicalNameHasQualifier("ROOT/x(q)")) // true
	fmt.Println(logicalnames.LogicalNameIsArtifact("ARTIFACT/x/y")) // true
}
```
