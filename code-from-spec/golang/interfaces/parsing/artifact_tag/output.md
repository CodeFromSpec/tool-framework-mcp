[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@n3vlZYg8oEtFLOA_I36eXUzQseQ)

# Package `artifacttag`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"`

## Package Declaration

```go
package artifacttag
```

## Struct Definitions

```go
package artifacttag

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
