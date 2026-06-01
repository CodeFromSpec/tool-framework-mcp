[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@sHhR71DnfK5pEcG0vFHz_VY02tA)

# Package `logicalnames`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
```

Package `logicalnames` provides functions for working with logical names — the `ROOT/` and `ARTIFACT/` reference strings used throughout the framework to identify spec nodes and generated artifacts.

---

## Error Sentinels

```go
package logicalnames

import "errors"

// ErrUnsupportedReference is returned when a logical name is not a
// ROOT/ reference (neither ROOT nor ROOT/...).
var ErrUnsupportedReference = errors.New("unsupported reference: expected a ROOT/ reference")

// ErrInvalidPath is returned when a path is not a _node.md file
// under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and
// has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent node")

// ErrNotARootReference is returned when the logical name is not a
// ROOT/ reference (neither ROOT nor ROOT/...).
var ErrNotARootReference = errors.New("not a ROOT/ reference")

// ErrNotAnArtifactReference is returned when the logical name does
// not start with ARTIFACT/.
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")
```

---

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of
// the corresponding _node.md file. Strips any qualifier before
// resolving. Only accepts ROOT/ references (including ROOT itself).
//
// Errors:
//   - ErrUnsupportedReference: the logical name is not a ROOT/
//     reference (neither ROOT nor ROOT/...).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
//
// Errors:
//   - ErrInvalidPath: the path is not a _node.md file under
//     code-from-spec/.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/
// references (including ROOT itself, which returns ErrNoParent).
//
// Errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name is not a ROOT/ reference
//     (neither ROOT nor ROOT/...).
func LogicalNameGetParent(logicalName string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns ("", false) if no qualifier is present. Works
// with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"        → ("z", true)
//   - "ROOT/x/y"           → ("", false)
//   - "ARTIFACT/x/y(id)"   → ("id", true)
func LogicalNameGetQualifier(logicalName string) (qualifier string, present bool)

// LogicalNameStripQualifier returns the logical name without the
// parenthetical qualifier. If no qualifier is present, returns the
// input unchanged. Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - "ROOT/x/y(z)"        → "ROOT/x/y"
//   - "ARTIFACT/x/y(id)"   → "ARTIFACT/x/y"
//   - "ROOT/x/y"           → "ROOT/x/y"
func LogicalNameStripQualifier(logicalName string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/
// reference other than ROOT itself. Returns false for ROOT, ARTIFACT/
// references, and unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/
// references.
func LogicalNameHasQualifier(logicalName string) bool

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of
// the node that generates the referenced artifact. Strips the
// ARTIFACT/ prefix and any qualifier.
//
// Examples:
//   - "ARTIFACT/x/y(id)"  → "ROOT/x/y"
//   - "ARTIFACT/x/y"      → "ROOT/x/y"
//
// Errors:
//   - ErrNotAnArtifactReference: the logical name does not start with
//     ARTIFACT/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	// Convert a ROOT/ logical name to a _node.md PathCfs.
	cfsPath, err := logicalnames.LogicalNameToPath("ROOT/x/y")
	if err != nil {
		if errors.Is(err, logicalnames.ErrUnsupportedReference) {
			log.Fatal("expected a ROOT/ reference")
		}
		log.Fatalf("failed to convert logical name to path: %v", err)
	}
	fmt.Println("node path:", cfsPath.Value) // code-from-spec/x/y/_node.md

	// Derive a ROOT/ logical name from a _node.md PathCfs.
	nodePath := &pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"}
	name, err := logicalnames.LogicalNameFromPath(nodePath)
	if err != nil {
		if errors.Is(err, logicalnames.ErrInvalidPath) {
			log.Fatal("path is not a _node.md file under code-from-spec/")
		}
		log.Fatalf("failed to derive logical name: %v", err)
	}
	fmt.Println("logical name:", name) // ROOT/x/y

	// Get the parent of a logical name.
	parent, err := logicalnames.LogicalNameGetParent("ROOT/x/y")
	if err != nil {
		if errors.Is(err, logicalnames.ErrNoParent) {
			log.Fatal("ROOT has no parent")
		}
		if errors.Is(err, logicalnames.ErrNotARootReference) {
			log.Fatal("expected a ROOT/ reference")
		}
		log.Fatalf("failed to get parent: %v", err)
	}
	fmt.Println("parent:", parent) // ROOT/x

	// Extract and strip a qualifier.
	qualified := "ROOT/x/y(z)"
	qualifier, present := logicalnames.LogicalNameGetQualifier(qualified)
	if present {
		fmt.Println("qualifier:", qualifier) // z
	}
	stripped := logicalnames.LogicalNameStripQualifier(qualified)
	fmt.Println("stripped:", stripped) // ROOT/x/y

	// Check properties of a logical name.
	fmt.Println("has parent:", logicalnames.LogicalNameHasParent("ROOT/x/y"))    // true
	fmt.Println("has parent:", logicalnames.LogicalNameHasParent("ROOT"))        // false
	fmt.Println("has qualifier:", logicalnames.LogicalNameHasQualifier(qualified)) // true
	fmt.Println("is artifact:", logicalnames.LogicalNameIsArtifact("ARTIFACT/x/y(id)")) // true

	// Get the generator node for an artifact reference.
	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y(id)")
	if err != nil {
		if errors.Is(err, logicalnames.ErrNotAnArtifactReference) {
			log.Fatal("expected an ARTIFACT/ reference")
		}
		log.Fatalf("failed to get artifact generator: %v", err)
	}
	fmt.Println("generator:", generator) // ROOT/x/y
}
```
