[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@fOPEHJNTORhh8dgPdfn1lJuHpwM)

# Package `mcpwritefile`

```go
package mcpwritefile
```

Import path: `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile`

## Error Sentinels

```go
package mcpwritefile

import "errors"

var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")
var ErrNoOutput              = errors.New("no output")
var ErrPathNotInOutput       = errors.New("path not in output")
```

## Functions

```go
package mcpwritefile

// MCPWriteFile validates that path is declared in the output field of the node
// identified by logical_name, then writes content to that path. Returns a
// success message of the form "wrote <path>" on success.
//
// Returns ErrUnreadableFrontmatter if the node's frontmatter cannot be parsed,
// ErrNoOutput if the target node has no output field, ErrPathNotInOutput if
// path is not declared in the node's output, or propagated errors from
// LogicalNames, PathUtils, and FileWriter packages.
func MCPWriteFile(logical_name string, path string, content string) (string, error)
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
		"# Package `mcpwritefile`\n",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```
