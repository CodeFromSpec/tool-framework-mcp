[//]: # (code-from-spec: SPEC/golang/interfaces/parsing/artifact_tag@SPJnnLR14cRdcHI89w73DqYA1OA)

# Package `artifacttag`

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v4/internal/artifacttag`

## Types

```go
package artifacttag

// ArtifactTag holds the parsed contents of a code-from-spec tag
// found inside a generated file.
type ArtifactTag struct {
	LogicalName string
	Hash        string
}
```

## Error Sentinels

```go
package artifacttag

import "errors"

var ErrNoTagFound   = errors.New("no code-from-spec tag found in file")
var ErrMalformedTag = errors.New("code-from-spec tag is malformed")
```

## Functions

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/pathutils"

// ArtifactTagExtract scans the file at file_path line by line, looking for
// a "code-from-spec: <logical-name>@<hash>" pattern. It returns the first
// match parsed into an ArtifactTag. Returns ErrNoTagFound if no line
// contains the pattern, or ErrMalformedTag if the pattern is present but
// cannot be parsed (missing "@", empty name, or wrong hash length).
// File-related errors from opening or reading the file are propagated as-is.
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
	filePath := &pathutils.PathCfs{Value: "code-from-spec/functional/payments/fees/output.md"}

	tag, err := artifacttag.ArtifactTagExtract(filePath)
	if err != nil {
		if errors.Is(err, artifacttag.ErrNoTagFound) {
			fmt.Println("File has not been generated from a spec.")
			return
		}
		if errors.Is(err, artifacttag.ErrMalformedTag) {
			fmt.Println("File contains a malformed artifact tag.")
			return
		}
		log.Fatal(err)
	}

	fmt.Printf("Logical name: %s\n", tag.LogicalName)
	fmt.Printf("Hash:         %s\n", tag.Hash)
}
```
