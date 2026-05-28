[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@axzN9uSE7z30JhEmKNZBEH3ekfg)

# Interface: `artifacttag`

## Package

```go
package artifacttag
```

## Import

```go
import "github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
```

---

## Struct Definitions

```go
// ArtifactTag holds the parsed contents of a code-from-spec tag
// found inside a generated file.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Example tag line:
//
//	// code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@axzN9uSE7z30JhEmKNZBEH3ekfg
type ArtifactTag struct {
	// LogicalName is the logical name component of the tag (e.g.
	// "ROOT/golang/interfaces/parsing/artifact_tag").
	LogicalName string

	// Hash is the hash component of the tag (e.g.
	// "axzN9uSE7z30JhEmKNZBEH3ekfg").
	Hash string
}
```

---

## Error Sentinels

```go
var (
	// ErrNoTagFound is returned when the file contains no
	// "code-from-spec:" substring on any line.
	ErrNoTagFound = errors.New("no tag found")

	// ErrMalformedTag is returned when the tag string is present but
	// cannot be parsed — for example, the "@" separator is missing,
	// the logical name is empty, or the hash has the wrong length.
	ErrMalformedTag = errors.New("malformed tag")
)
```

---

## Functions

```go
// ArtifactTagExtract opens the file at file_path, scans its lines for
// the first occurrence of the "code-from-spec:" pattern, and returns
// the parsed ArtifactTag.
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Comment syntax is not parsed — every line is scanned for the substring
// "code-from-spec:" regardless of context.
//
// Possible errors:
//   - Path errors propagated from opening the file (e.g. pathutils.ErrPathEmpty,
//     pathutils.ErrPathAbsolute, pathutils.ErrPathContainsBackslash,
//     pathutils.ErrDirectoryTraversal, pathutils.ErrResolvesOutsideRoot,
//     pathutils.ErrCannotDetermineRoot).
//   - ErrNoTagFound — the file was read successfully but contains no tag.
//   - ErrMalformedTag — a tag line was found but could not be parsed.
func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error)
```

---

## Usage Examples

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v2/internal/pathutils"
)

func main() {
	// Build a CFS path pointing to the file to inspect.
	cfs := &pathutils.PathCfs{Value: "internal/artifacttag/artifacttag.go"}

	// Extract the artifact tag from the file.
	tag, err := artifacttag.ArtifactTagExtract(cfs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Logical name:", tag.LogicalName)
	fmt.Println("Hash:", tag.Hash)
}
```
