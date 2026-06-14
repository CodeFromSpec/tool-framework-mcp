[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@J35rln9Cz1AXRMO56qqD7A5ZvJY)

# Package `artifacttag`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag`

---

## Structs

```go
package artifacttag

// ArtifactTag holds the parsed components of a code-from-spec tag
// found inside a generated source file.
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

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrNoTagFound     = errors.New("file has no code-from-spec: tag")
var ErrMalformedTag   = errors.New("tag exists but cannot be parsed")
```

---

## Functions

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ArtifactTagExtract opens the file at filePath, scans each line for a
// code-from-spec: tag, and returns the parsed ArtifactTag.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// The tag may appear inside any comment syntax. Only the first occurrence
// is used. Returns ErrFileUnreadable if the file cannot be opened or read,
// ErrNoTagFound if no tag is present, and ErrMalformedTag if the tag
// exists but lacks the @ separator, has an empty logical name, or has a
// hash of the wrong length.
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
	path := &pathutils.PathCfs{Value: "ARTIFACT/functional/fees/output.go"}

	tag, err := artifacttag.ArtifactTagExtract(path)
	if err != nil {
		if errors.Is(err, artifacttag.ErrNoTagFound) {
			fmt.Println("file has not been generated from a spec")
			return
		}
		if errors.Is(err, artifacttag.ErrMalformedTag) {
			fmt.Println("file contains a malformed tag")
			return
		}
		log.Fatal(err)
	}

	fmt.Println("Logical name:", tag.LogicalName)
	fmt.Println("Hash:        ", tag.Hash)
}
```
