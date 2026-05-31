[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@iWDRoM0HVB3AZC7pQ0mGBOOC0Ls)

# Package `artifacttag`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
```

Provides utilities to extract and parse the `code-from-spec:` artifact tag embedded in generated source files.

---

## Structs

```go
package artifacttag

// ArtifactTag holds the parsed components of a code-from-spec tag found in a file.
//
// The tag has the format:
//
//	code-from-spec: <logical-name>@<hash>
//
// It may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
type ArtifactTag struct {
	LogicalName string
	Hash        string
}
```

---

## Error Sentinels

```go
package artifacttag

import "errors"

// ErrFileUnreadable is returned when the file cannot be opened or read.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrNoTagFound is returned when the file contains no code-from-spec: substring.
var ErrNoTagFound = errors.New("no tag found")

// ErrMalformedTag is returned when a code-from-spec: tag exists but cannot be
// parsed — missing @, empty logical name, or wrong hash length.
var ErrMalformedTag = errors.New("malformed tag")
```

---

## Functions

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ArtifactTagExtract opens the file at filePath, scans each line for the
// code-from-spec: pattern, and returns the parsed ArtifactTag.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Comment syntax is ignored — any line containing the substring is considered.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrNoTagFound: the file has no code-from-spec: substring.
//   - ErrMalformedTag: the tag exists but cannot be parsed (no @, empty name,
//     wrong hash length).
//   - (FileReader.*): propagated from FileOpen.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	// Extract the artifact tag from a generated source file.
	cfs := &pathutils.PathCfs{Value: "internal/artifacttag/artifacttag.go"}
	tag, err := artifacttag.ArtifactTagExtract(cfs)
	if err != nil {
		switch err {
		case artifacttag.ErrFileUnreadable:
			log.Fatalf("could not read file: %v", err)
		case artifacttag.ErrNoTagFound:
			log.Fatalf("file has no artifact tag: %v", err)
		case artifacttag.ErrMalformedTag:
			log.Fatalf("artifact tag is malformed: %v", err)
		default:
			log.Fatalf("unexpected error: %v", err)
		}
	}

	fmt.Println("Logical name:", tag.LogicalName)
	fmt.Println("Hash:        ", tag.Hash)
}
```
