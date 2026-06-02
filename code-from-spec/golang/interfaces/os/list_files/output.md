[//]: # (code-from-spec: ROOT/golang/interfaces/os/list_files@YFY0LKmxwzxvqRTn9pQK5fDzzlw)

# Package `listfiles`

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"`

## Package Declaration

```go
package listfiles
```

## Error Sentinels

```go
package listfiles

import "errors"

var ErrDirectoryNotFound = errors.New("directory not found")
var ErrWalkError         = errors.New("filesystem walk error")
```

## Function Signatures

```go
package listfiles

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"

func ListFiles(cfs_path *pathutils.PathCfs) ([]*pathutils.PathCfs, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/listfiles"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/pathutils"
)

func main() {
	dir := &pathutils.PathCfs{Value: "code-from-spec"}

	files, err := listfiles.ListFiles(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Value)
	}
}
```
