[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@s9VIuXEVdYf5c4OCPW029VCSIdc)

# Package `artifacttag`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"`

## Package Declaration

```go
package artifacttag
```

## Struct Definitions

```go
package artifacttag

// ArtifactTag holds the parsed contents of a code-from-spec tag found
// inside a generated file.
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
var ErrNoTagFound     = errors.New("no code-from-spec tag found")
var ErrMalformedTag   = errors.New("tag is malformed")
```

## Function Signatures

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ArtifactTagExtract scans the file at file_path line by line for a
// code-from-spec: <logical-name>@<hash> tag. The tag may appear inside
// any comment syntax. Returns the parsed ArtifactTag on success.
func ArtifactTagExtract(file_path *pathutils.PathCfs) (*ArtifactTag, error)
```

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
	file := &pathutils.PathCfs{Value: "internal/filereader/filereader.go"}

	tag, err := artifacttag.ArtifactTagExtract(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("logical name:", tag.LogicalName)
	fmt.Println("hash:", tag.Hash)
}
```
