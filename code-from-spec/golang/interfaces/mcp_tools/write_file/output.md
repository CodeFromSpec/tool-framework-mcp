[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@9GZb3_qHKmY5a23idc1n8wurS4E)

# Package `mcpwritefile`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
```

## Error Sentinels

```go
package mcpwritefile

import "errors"

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput = errors.New("target node has no output field")
var ErrPathNotInOutput = errors.New("path is not declared in the node's output")
```

## Functions

```go
package mcpwritefile

// MCPWriteFile writes content to the file at path, authorized by the node
// identified by logical_name. It validates that path is declared in the
// node's output frontmatter field before writing. Returns "wrote <path>"
// on success.
func MCPWriteFile(logicalName string, path string, content string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
)

func main() {
	result, err := mcpwritefile.MCPWriteFile(
		"ROOT/golang/interfaces/mcp_tools/write_file",
		"code-from-spec/golang/interfaces/mcp_tools/write_file/output.md",
		"# Output\n\nGenerated content.\n",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
```
