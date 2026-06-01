[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@Na2fdUmffqbI_YdC0liSgTl_-fQ)

# Package `artifacttag`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
```

Package `artifacttag` provides functionality for extracting the `code-from-spec` artifact tag from generated source files. The tag encodes the logical name and hash of the spec that produced the file.

---

## Structs

```go
package artifacttag

// ArtifactTag holds the parsed components of a code-from-spec tag
// found in a generated source file.
//
// Tag format:
//
//	code-from-spec: <logical-name>@<hash>
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Parsing is line-based and does not interpret comment delimiters.
type ArtifactTag struct {
	// LogicalName is the logical node name extracted from the tag,
	// for example "ROOT/golang/interfaces/parsing/artifact_tag".
	LogicalName string

	// Hash is the chain hash extracted from the tag,
	// for example "Na2fdUmffqbI_YdC0liSgTl_-fQ".
	Hash string
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

// ErrMalformedTag is returned when the tag exists but cannot be parsed
// (missing @, empty logical name, or wrong hash length).
var ErrMalformedTag = errors.New("malformed tag")
```

---

## Functions

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ArtifactTagExtract scans the file at filePath line by line for the
// first occurrence of the substring "code-from-spec:" and parses the
// logical name and hash from it.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Parsing is purely textual — comment delimiters are ignored.
//
// Errors:
//   - ErrFileUnreadable: the file cannot be opened or read.
//   - ErrNoTagFound: no "code-from-spec:" substring was found in the file.
//   - ErrMalformedTag: the tag was found but could not be parsed
//     (e.g. no @ separator, empty logical name, or wrong hash length).
//   - (FileReader.*): propagated from FileOpen.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	// Build a CFS path to a generated source file.
	cfsPath := &pathutils.PathCfs{Value: "internal/mypackage/myfile.go"}

	// Extract the artifact tag from the file.
	tag, err := artifacttag.ArtifactTagExtract(cfsPath)
	if err != nil {
		if errors.Is(err, artifacttag.ErrNoTagFound) {
			log.Fatal("file has no code-from-spec tag — it may not be generated")
		}
		if errors.Is(err, artifacttag.ErrMalformedTag) {
			log.Fatal("code-from-spec tag exists but is malformed")
		}
		if errors.Is(err, artifacttag.ErrFileUnreadable) {
			log.Fatal("could not read the file")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Printf("logical name : %s\n", tag.LogicalName)
	fmt.Printf("hash         : %s\n", tag.Hash)
}
```
