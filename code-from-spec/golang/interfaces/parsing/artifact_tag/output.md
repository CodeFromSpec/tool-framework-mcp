[//]: # (code-from-spec: SPEC/golang/interfaces/parsing/artifact_tag@AadiSpuWQgO-90Y3EuSYsujR5aA)

# Package `artifacttag`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag`

## Struct Definitions

```go
package artifacttag

// ArtifactTag holds the parsed contents of a code-from-spec tag found in a
// generated file.
type ArtifactTag struct {
	LogicalName string
	Hash        string
}
```

## Error Sentinels

```go
package artifacttag

import "errors"

var ErrFileUnreadable = errors.New("file cannot be opened or read")
var ErrNoTagFound     = errors.New("file has no code-from-spec: tag")
var ErrMalformedTag   = errors.New("tag exists but cannot be parsed")
```

## Function Signatures

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// ArtifactTagExtract opens the file at filePath and scans its lines for a
// code-from-spec: <logical-name>@<hash> tag. The tag may appear inside any
// comment syntax; the function does not parse comment delimiters.
//
// Returns ErrFileUnreadable if the file cannot be opened or read.
// Returns ErrNoTagFound if no line contains the "code-from-spec:" substring.
// Returns ErrMalformedTag if the tag is present but has no "@", an empty
// logical name, or a hash of unexpected length.
// Errors from the underlying file reader are propagated as-is.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error)
```

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"
)

func main() {
	cfsPath := &pathutils.PathCfs{Value: "src/payments/fees.go"}

	tag, err := artifacttag.ArtifactTagExtract(cfsPath)
	if err != nil {
		if errors.Is(err, artifacttag.ErrNoTagFound) {
			fmt.Println("file has not been generated from a spec")
			return
		}
		if errors.Is(err, artifacttag.ErrMalformedTag) {
			fmt.Println("tag is present but malformed")
			return
		}
		log.Fatal(err)
	}

	fmt.Printf("logical name : %s\n", tag.LogicalName)
	fmt.Printf("hash         : %s\n", tag.Hash)
}
```
