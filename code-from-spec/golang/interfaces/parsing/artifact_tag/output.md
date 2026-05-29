[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@g1QvX6BKXj44-GjQn0opXX3yXbE)

# Interface: `artifacttag`

**Package:** `package artifacttag`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"`

---

## Structs

```go
// ArtifactTag holds the parsed contents of a code-from-spec tag
// found in a generated source file.
//
// The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// Example tag (inside a comment):
//
//	// code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@g1QvX6BKXj44-GjQn0opXX3yXbE
type ArtifactTag struct {
    LogicalName string
    Hash        string
}
```

---

## Error Sentinels

```go
var (
    // ErrFileUnreadable is returned when the file cannot be opened or read.
    ErrFileUnreadable = errors.New("file unreadable")

    // ErrNoTagFound is returned when the file contains no code-from-spec: substring.
    ErrNoTagFound = errors.New("no tag found")

    // ErrMalformedTag is returned when a code-from-spec: substring is found
    // but cannot be fully parsed (missing "@", empty logical name, or wrong
    // hash length).
    ErrMalformedTag = errors.New("malformed tag")
)
```

---

## Functions

```go
// ArtifactTagExtract opens the file at file_path, scans each line for the
// pattern "code-from-spec: <logical-name>@<hash>", and returns the parsed
// tag on the first match.
//
// The tag may appear inside any comment syntax (//, #, /* */, --, <!-- -->).
// Comment syntax is not parsed — the function scans raw lines for the
// substring "code-from-spec:".
//
// Returns:
//   - (*ArtifactTag, nil) on success.
//   - (nil, ErrFileUnreadable) if the file cannot be opened or read.
//   - (nil, ErrNoTagFound) if no line contains "code-from-spec:".
//   - (nil, ErrMalformedTag) if the tag exists but cannot be parsed
//     (no "@" separator, empty logical name, or wrong hash length).
//
// Path errors from opening the file are propagated directly.
func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error)
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
    // Point at a generated source file that contains a code-from-spec tag.
    filePath := &pathutils.PathCfs{Value: "internal/artifacttag/artifacttag.go"}

    tag, err := artifacttag.ArtifactTagExtract(filePath)
    if err != nil {
        log.Fatalf("could not extract artifact tag: %v", err)
    }

    fmt.Println("Logical name:", tag.LogicalName)
    fmt.Println("Hash:        ", tag.Hash)
}
```
