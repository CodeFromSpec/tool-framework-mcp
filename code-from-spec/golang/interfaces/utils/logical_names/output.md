[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@yCpVo8tBxZmG4nw3WlL9WGh44Gw)

# Package `logicalnames`

```go
package logicalnames
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames`

## Error Sentinels

```go
package logicalnames

import "errors"

var ErrUnsupportedReference   = errors.New("unsupported reference")
var ErrInvalidPath            = errors.New("invalid path")
var ErrNoParent               = errors.New("no parent")
var ErrNotARootReference      = errors.New("not a ROOT reference")
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT reference")
```

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references (including ROOT itself).
// Returns ErrUnsupportedReference if the logical name is not a ROOT/ reference.
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file path.
// The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
// Returns ErrInvalidPath if the path is not a _node.md file under code-from-spec/.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
// Only accepts ROOT/ references (including ROOT itself).
// Returns ErrNoParent if the logical name is ROOT itself.
// Returns ErrNotARootReference if the logical name is not a ROOT/ reference.
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Returns an empty string and false if no qualifier is present.
// Works with both ROOT/ and ARTIFACT/ references; ARTIFACT/ references always
// return ("", false).
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with both ROOT/ and ARTIFACT/ references.
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself. Returns false for ROOT, ARTIFACT/ references, and
// unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with both ROOT/ and ARTIFACT/ references;
// ARTIFACT/ references always return false.
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix.
// Returns ErrNotAnArtifactReference if the logical name does not start with ARTIFACT/.
func LogicalNameGetArtifactGenerator(logical_name string) (string, error)
```

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
	path, err := logicalnames.LogicalNameToPath("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("path:", path.Value)
	// path: code-from-spec/x/y/_node.md

	name, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("logical name:", name)
	// logical name: ROOT/x/y

	parent, err := logicalnames.LogicalNameGetParent("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("parent:", parent)
	// parent: ROOT/x

	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(z)")
	fmt.Println("qualifier:", qualifier, "present:", ok)
	// qualifier: z present: true

	stripped := logicalnames.LogicalNameStripQualifier("ROOT/x/y(z)")
	fmt.Println("stripped:", stripped)
	// stripped: ROOT/x/y

	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("generator:", generator)
	// generator: ROOT/x/y
}
```
