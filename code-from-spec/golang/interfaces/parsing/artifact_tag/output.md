[//]: # (code-from-spec: ROOT/golang/interfaces/parsing/artifact_tag@2MoMJHo8CjlhjeL-0_j6bt5izp4)

# Package `artifacttag`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/artifacttag"
```

## Structs

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
var ErrNoTagFound = errors.New("file has no code-from-spec: substring")
var ErrMalformedTag = errors.New("tag exists but cannot be parsed")
```

## Functions

```go
package artifacttag

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

// ArtifactTagExtract scans the file at file_path for a code-from-spec tag
// and returns the parsed ArtifactTag. The tag format is:
//
//	code-from-spec: <logical-name>@<hash>
//
// The tag may appear inside any comment syntax. Each line is scanned
// for the pattern regardless of context.
func ArtifactTagExtract(filePath *pathutils.PathCfs) (*ArtifactTag, error)
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
	cfsPath := &pathutils.PathCfs{Value: "internal/mypackage/myfile.go"}

	tag, err := artifacttag.ArtifactTagExtract(cfsPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Logical name:", tag.LogicalName)
	fmt.Println("Hash:", tag.Hash)
}
```
