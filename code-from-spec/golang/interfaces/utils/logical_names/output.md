[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@5IFac7_60GtUB7jtbZvp04dlYLA)

# Package `logicalnames`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames"
```

## Structs

None.

## Error Sentinels

```go
package logicalnames

import "errors"

var ErrUnsupportedReference = errors.New("unsupported reference")
var ErrInvalidPath = errors.New("invalid path")
var ErrNoParent = errors.New("no parent")
var ErrNotARootReference = errors.New("not a ROOT/ reference")
var ErrNotAnArtifactReference = errors.New("not an ARTIFACT/ reference")
```

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// LogicalNameToPath converts a ROOT/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts ROOT/ references (including ROOT itself).
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the ROOT/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a ROOT/ reference.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent. Only accepts ROOT/
// references (including ROOT itself, which returns ErrNoParent).
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical
// name. Returns empty string and false if no qualifier is present.
func LogicalNameGetQualifier(logical_name string) (string, bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a ROOT/ reference
// other than ROOT itself.
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier.
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the ROOT/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix.
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
	cfsPath, err := logicalnames.LogicalNameToPath("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Path:", cfsPath.Value)

	name, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", name)

	parent, err := logicalnames.LogicalNameGetParent("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent)

	qualifier, hasQ := logicalnames.LogicalNameGetQualifier("ROOT/x/y(z)")
	if hasQ {
		fmt.Println("Qualifier:", qualifier)
	}

	stripped := logicalnames.LogicalNameStripQualifier("ROOT/x/y(z)")
	fmt.Println("Stripped:", stripped)

	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator:", generator)
}
```
