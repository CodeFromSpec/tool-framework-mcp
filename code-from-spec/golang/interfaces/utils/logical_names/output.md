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
// Possible errors:
//   - ErrUnsupportedReference: the logical name does not start with ROOT/.
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md
// file path. The inverse of LogicalNameToPath. Always returns a ROOT/
// reference.
//
// Possible errors:
//   - ErrInvalidPath: the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/
// references.
//
// Possible errors:
//   - ErrNoParent: the logical name is ROOT itself.
//   - ErrNotARootReference: the logical name does not start with ROOT/.
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a
// logical name. Returns ("", false) if no qualifier is present. Works
// with both ROOT/ and ARTIFACT/ references.
//
// For example:
//   - ROOT/x/y(z)  → ("z", true)
//   - ROOT/x/y     → ("", false)
func LogicalNameGetQualifier(logical_name string) (string, bool)

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
// the node that generates the referenced artifact. Strips the ARTIFACT/
// prefix and any qualifier.
//
// For example:
//   - ARTIFACT/x/y(id) → ROOT/x/y
//   - ARTIFACT/x/y     → ROOT/x/y
//
// Possible errors:
//   - ErrNotAnArtifactReference: the logical name does not start with ARTIFACT/.
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
	cfs, err := logicalnames.LogicalNameToPath("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Path:", cfs.Value) // code-from-spec/x/y/_node.md

	// Convert a PathCfs back to a ROOT/ logical name.
	name, err := logicalnames.LogicalNameFromPath(cfs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", name) // ROOT/x/y

	// Get the parent of a logical name.
	parent, err := logicalnames.LogicalNameGetParent("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent) // ROOT/x

	// Extract a qualifier.
	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(z)")
	if ok {
		fmt.Println("Qualifier:", qualifier) // z
	}

	// Check for a qualifier without extracting it.
	fmt.Println("Has qualifier:", logicalnames.LogicalNameHasQualifier("ROOT/x/y(z)")) // true
	fmt.Println("Has qualifier:", logicalnames.LogicalNameHasQualifier("ROOT/x/y"))    // false

	// Check parent navigation.
	fmt.Println("Has parent:", logicalnames.LogicalNameHasParent("ROOT/x/y")) // true
	fmt.Println("Has parent:", logicalnames.LogicalNameHasParent("ROOT"))     // false

	// Check if a logical name is an artifact reference.
	fmt.Println("Is artifact:", logicalnames.LogicalNameIsArtifact("ARTIFACT/x/y(id)")) // true
	fmt.Println("Is artifact:", logicalnames.LogicalNameIsArtifact("ROOT/x/y"))         // false

	// Resolve an artifact reference to its generating node.
	// Step 1: get the generator's logical name.
	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y(id)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator:", generator) // ROOT/x/y

	// Step 2: get the generator node's path.
	genPath, err := logicalnames.LogicalNameToPath(generator)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator path:", genPath.Value) // code-from-spec/x/y/_node.md

	// Step 3: get the artifact id.
	artifactID, _ := logicalnames.LogicalNameGetQualifier("ARTIFACT/x/y(id)")
	fmt.Println("Artifact id:", artifactID) // id

	// Step 4: read the node's frontmatter, find the output entry whose
	// id matches artifactID, and use its path to locate the artifact file.
	_ = artifactID
	_ = genPath
}
```
