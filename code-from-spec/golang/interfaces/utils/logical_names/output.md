[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@_l3gjuceHK6LGvtlnfdjVWC54eY)

# Package `logicalnames`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
```

Provides utilities for parsing, navigating, and converting logical names used throughout the Code from Spec framework. Logical names identify spec nodes (`ROOT/`) and generated artifacts (`ARTIFACT/`) and may carry an optional parenthetical qualifier.

---

## Error Sentinels

```go
package logicalnames

import "errors"

// ErrUnsupportedReference is returned when a logical name does not start with ROOT/.
var ErrUnsupportedReference = errors.New("unsupported reference: logical name must start with ROOT/")

// ErrInvalidPath is returned when a path is not a _node.md file under code-from-spec/.
var ErrInvalidPath = errors.New("invalid path: not a _node.md file under code-from-spec/")

// ErrNoParent is returned when the logical name is ROOT itself and has no parent.
var ErrNoParent = errors.New("no parent: ROOT has no parent node")

// ErrNotARootReference is returned when the logical name does not start with ROOT/.
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
// Only accepts ROOT/ references.
//
// Examples:
//   - ROOT              → code-from-spec/_node.md
//   - ROOT/x/y          → code-from-spec/x/y/_node.md
//   - ROOT/x/y(z)       → code-from-spec/x/y/_node.md
//
// Errors:
//   - ErrUnsupportedReference: the logical name does not start with ROOT/.
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file path.
// The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
//
// Examples:
//   - code-from-spec/_node.md       → ROOT
//   - code-from-spec/x/_node.md     → ROOT/x
//   - code-from-spec/x/y/_node.md   → ROOT/x/y
//
// Errors:
//   - ErrInvalidPath: the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/ references.
//
// Examples:
//   - ROOT/x      → ROOT
//   - ROOT/x/y    → ROOT/x
//   - ROOT/x/y(z) → ROOT/x
//
// Errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name does not start with ROOT/.
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Returns an empty string and false if no qualifier is present.
// Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - ROOT/x/y(z)      → "z", true
//   - ARTIFACT/x/y(id) → "id", true
//   - ROOT/x/y         → "", false
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with both ROOT/ and ARTIFACT/ references.
//
// Examples:
//   - ROOT/x/y(z)       → ROOT/x/y
//   - ARTIFACT/x/y(id)  → ARTIFACT/x/y
//   - ROOT/x/y          → ROOT/x/y
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself. Returns false for ROOT, ARTIFACT/ references, and
// unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and any
// qualifier.
//
// Examples:
//   - ARTIFACT/x/y(id) → ROOT/x/y
//   - ARTIFACT/x/y     → ROOT/x/y
//
// Errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
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
		log.Fatalf("LogicalNameToPath: %v", err)
	}
	fmt.Println("Node path:", cfsPath.Value)
	// Output: code-from-spec/golang/interfaces/utils/logical_names/_node.md

	// Derive the logical name back from a PathCfs.
	name, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/golang/interfaces/utils/logical_names/_node.md"})
	if err != nil {
		log.Fatalf("LogicalNameFromPath: %v", err)
	}
	fmt.Println("Logical name:", name)
	// Output: ROOT/golang/interfaces/utils/logical_names

	// Navigate to the parent node.
	parent, err := logicalnames.LogicalNameGetParent("ROOT/golang/interfaces/utils/logical_names")
	if err != nil {
		log.Fatalf("LogicalNameGetParent: %v", err)
	}
	fmt.Println("Parent:", parent)
	// Output: ROOT/golang/interfaces/utils

	// Check and extract a qualifier.
	qualified := "ROOT/golang/interfaces/utils/logical_names(interface)"
	if logicalnames.LogicalNameHasQualifier(qualified) {
		q, _ := logicalnames.LogicalNameGetQualifier(qualified)
		fmt.Println("Qualifier:", q)
		// Output: interface
	}
	stripped := logicalnames.LogicalNameStripQualifier(qualified)
	fmt.Println("Stripped:", stripped)
	// Output: ROOT/golang/interfaces/utils/logical_names

	// Resolve an ARTIFACT/ reference to its generating node.
	artifactRef := "ARTIFACT/golang/interfaces/utils/logical_names(interface)"
	if logicalnames.LogicalNameIsArtifact(artifactRef) {
		generator, err := logicalnames.LogicalNameGetArtifactGenerator(artifactRef)
		if err != nil {
			log.Fatalf("LogicalNameGetArtifactGenerator: %v", err)
		}
		fmt.Println("Generator node:", generator)
		// Output: ROOT/golang/interfaces/utils/logical_names
	}
}
```
