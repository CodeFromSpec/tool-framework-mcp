[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@fa2bctneLu0eedTuyJ4J3WB8yNI)

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

func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)
func LogicalNameGetParent(logical_name string) (string, error)
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool)
func LogicalNameStripQualifier(logical_name string) string
func LogicalNameHasParent(logical_name string) bool
func LogicalNameHasQualifier(logical_name string) bool
func LogicalNameIsArtifact(logical_name string) bool
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

	name, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/x/y/_node.md"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("logical name:", name)

	parent, err := logicalnames.LogicalNameGetParent("ROOT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("parent:", parent)

	qualifier, ok := logicalnames.LogicalNameGetQualifier("ROOT/x/y(z)")
	fmt.Println("qualifier:", qualifier, "present:", ok)

	stripped := logicalnames.LogicalNameStripQualifier("ROOT/x/y(z)")
	fmt.Println("stripped:", stripped)

	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/x/y")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("generator:", generator)
}
```
