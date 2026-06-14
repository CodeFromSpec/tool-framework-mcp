[//]: # (code-from-spec: ROOT/golang/interfaces/utils/logical_names@Y6B3Gef4xT6QVEsioPaWuNWq-dk)

# Package `logicalnames`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/logicalnames`

---

## Error Sentinels

```go
package logicalnames

import "errors"

var ErrUnsupportedReference    = errors.New("unsupported reference: not a SPEC/ reference")
var ErrInvalidPath             = errors.New("invalid path: not a _node.md file under code-from-spec/")
var ErrNoParent                = errors.New("no parent: logical name is SPEC itself")
var ErrNotASpecReference       = errors.New("not a SPEC/ reference")
var ErrNotAnArtifactReference  = errors.New("not an ARTIFACT/ reference")
var ErrNotAnExternalReference  = errors.New("not an EXTERNAL/ reference")
```

---

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// LogicalNameToPath converts a SPEC/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts SPEC/ references (including SPEC itself).
func LogicalNameToPath(logical_name string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the SPEC/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a SPEC/ reference.
func LogicalNameFromPath(cfs_path *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node. Strips
// any qualifier before computing the parent. Only accepts SPEC/ references
// (including SPEC itself, which returns ErrNoParent). Always returns a
// SPEC/ reference.
func LogicalNameGetParent(logical_name string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical
// name. Returns an empty string and false if no qualifier is present. Works
// with SPEC/, ARTIFACT/, and EXTERNAL/ references; the latter two always
// return absent.
func LogicalNameGetQualifier(logical_name string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged. Works
// with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameStripQualifier(logical_name string) string

// LogicalNameHasParent returns true if the logical name is a SPEC/ reference
// other than SPEC itself. Returns false for SPEC, ARTIFACT/, EXTERNAL/, and
// unrecognized prefixes.
func LogicalNameHasParent(logical_name string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with SPEC/, ARTIFACT/, and EXTERNAL/
// references.
func LogicalNameHasQualifier(logical_name string) bool

// LogicalNameIsArtifact returns true if the logical name starts with
// ARTIFACT/.
func LogicalNameIsArtifact(logical_name string) bool

// LogicalNameIsSpec returns true if the logical name is exactly SPEC or
// starts with SPEC/.
func LogicalNameIsSpec(logical_name string) bool

// LogicalNameIsExternal returns true if the logical name starts with
// EXTERNAL/.
func LogicalNameIsExternal(logical_name string) bool

// LogicalNameGetArtifactGenerator returns the SPEC/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and
// prepends SPEC/. For example, ARTIFACT/x/y → SPEC/x/y.
func LogicalNameGetArtifactGenerator(logical_name string) (string, error)

// LogicalNameExternalToPath converts an EXTERNAL/ logical name to a PathCfs.
// Strips the EXTERNAL/ prefix and returns the remainder as a PathCfs
// relative to the project root.
// For example, EXTERNAL/proto/v1/api.proto → proto/v1/api.proto.
func LogicalNameExternalToPath(logical_name string) (*pathutils.PathCfs, error)
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
	path, err := logicalnames.LogicalNameToPath("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node path:", path.Value)

	name, err := logicalnames.LogicalNameFromPath(&pathutils.PathCfs{Value: "code-from-spec/payments/fees/_node.md"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", name)

	parent, err := logicalnames.LogicalNameGetParent("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent)

	qualifier, ok := logicalnames.LogicalNameGetQualifier("SPEC/payments/fees(Interface)")
	if ok {
		fmt.Println("Qualifier:", qualifier)
	}

	stripped := logicalnames.LogicalNameStripQualifier("SPEC/payments/fees(Interface)")
	fmt.Println("Stripped:", stripped)

	fmt.Println("Has parent:", logicalnames.LogicalNameHasParent("SPEC/payments/fees"))
	fmt.Println("Is spec:", logicalnames.LogicalNameIsSpec("SPEC/payments"))
	fmt.Println("Is artifact:", logicalnames.LogicalNameIsArtifact("ARTIFACT/payments/fees"))
	fmt.Println("Is external:", logicalnames.LogicalNameIsExternal("EXTERNAL/proto/api.proto"))

	generator, err := logicalnames.LogicalNameGetArtifactGenerator("ARTIFACT/payments/fees")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generator:", generator)

	extPath, err := logicalnames.LogicalNameExternalToPath("EXTERNAL/proto/v1/api.proto")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("External path:", extPath.Value)
}
```
