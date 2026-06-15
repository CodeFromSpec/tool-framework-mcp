[//]: # (code-from-spec: SPEC/golang/interfaces/utils/logical_names@5a-pu3StkoVjRspkoYcJy08Zo_s)

# Package `logicalnames`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames`

## Error Sentinels

```go
package logicalnames

import "errors"

var ErrUnsupportedReference   = errors.New("logical name is not a SPEC/ reference")
var ErrInvalidPath             = errors.New("path is not a _node.md file under code-from-spec/")
var ErrNoParent                = errors.New("logical name is SPEC itself")
var ErrNotASpecReference       = errors.New("logical name is not a SPEC/ reference")
var ErrNotAnArtifactReference  = errors.New("logical name does not start with ARTIFACT/")
var ErrNotAnExternalReference  = errors.New("logical name does not start with EXTERNAL/")
```

## Functions

```go
package logicalnames

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// LogicalNameToPath converts a SPEC/ logical name to the PathCfs of the
// corresponding _node.md file. Strips any qualifier before resolving.
// Only accepts SPEC/ references (including SPEC itself).
func LogicalNameToPath(logicalName string) (*pathutils.PathCfs, error)

// LogicalNameFromPath derives the SPEC/ logical name from a _node.md file
// path. The inverse of LogicalNameToPath. Always returns a SPEC/ reference.
func LogicalNameFromPath(cfsPath *pathutils.PathCfs) (string, error)

// LogicalNameGetParent returns the logical name of the parent node.
// Strips any qualifier before computing the parent.
// Only accepts SPEC/ references (including SPEC itself, which returns ErrNoParent).
// Always returns a SPEC/ reference.
func LogicalNameGetParent(logicalName string) (string, error)

// LogicalNameGetQualifier extracts the parenthetical qualifier from a logical name.
// Returns empty string and false if no qualifier is present.
// Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameGetQualifier(logicalName string) (qualifier string, ok bool)

// LogicalNameStripQualifier returns the logical name without the parenthetical
// qualifier. If no qualifier is present, returns the input unchanged.
// Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameStripQualifier(logicalName string) string

// LogicalNameHasParent returns true if the logical name is a SPEC/ reference
// other than SPEC itself. Returns false for SPEC, ARTIFACT/, EXTERNAL/,
// and unrecognized prefixes.
func LogicalNameHasParent(logicalName string) bool

// LogicalNameHasQualifier returns true if the logical name contains a
// parenthetical qualifier. Works with SPEC/, ARTIFACT/, and EXTERNAL/ references.
func LogicalNameHasQualifier(logicalName string) bool

// LogicalNameIsArtifact returns true if the logical name starts with ARTIFACT/.
func LogicalNameIsArtifact(logicalName string) bool

// LogicalNameIsSpec returns true if the logical name is exactly SPEC or
// starts with SPEC/.
func LogicalNameIsSpec(logicalName string) bool

// LogicalNameIsExternal returns true if the logical name starts with EXTERNAL/.
func LogicalNameIsExternal(logicalName string) bool

// LogicalNameGetArtifactGenerator returns the SPEC/ logical name of the node
// that generates the referenced artifact. Strips the ARTIFACT/ prefix and
// prepends SPEC/.
func LogicalNameGetArtifactGenerator(logicalName string) (string, error)

// LogicalNameExternalToPath converts an EXTERNAL/ logical name to a PathCfs.
// Strips the EXTERNAL/ prefix and returns the remainder as a PathCfs
// relative to the project root.
func LogicalNameExternalToPath(logicalName string) (*pathutils.PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	specName := "SPEC/payments/fees"

	nodePath, err := logicalnames.LogicalNameToPath(specName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node path:", nodePath.Value)

	roundTripped, err := logicalnames.LogicalNameFromPath(nodePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Logical name:", roundTripped)

	parent, err := logicalnames.LogicalNameGetParent(specName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parent:", parent)

	qualified := "SPEC/payments/fees(summary)"

	qualifier, ok := logicalnames.LogicalNameGetQualifier(qualified)
	if ok {
		fmt.Println("Qualifier:", qualifier)
	}

	stripped := logicalnames.LogicalNameStripQualifier(qualified)
	fmt.Println("Stripped:", stripped)

	fmt.Println("Has parent:", logicalnames.LogicalNameHasParent(specName))
	fmt.Println("Has qualifier:", logicalnames.LogicalNameHasQualifier(qualified))
	fmt.Println("Is spec:", logicalnames.LogicalNameIsSpec(specName))
	fmt.Println("Is artifact:", logicalnames.LogicalNameIsArtifact("ARTIFACT/payments/fees"))
	fmt.Println("Is external:", logicalnames.LogicalNameIsExternal("EXTERNAL/proto/v1/api.proto"))

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

	_ = &pathutils.PathCfs{}
}
```
