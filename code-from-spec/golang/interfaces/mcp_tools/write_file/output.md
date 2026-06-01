[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@rwbRTR9QLC9dw2_b0JQ7WYoCDJc)

# Package `mcpwritefile`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
```

Package `mcpwritefile` implements the MCP `write_file` tool. It validates that a given path is declared in the target node's outputs and writes the provided content to disk.

---

## Error Sentinels

```go
package mcpwritefile

import "errors"

// ErrUnreadableFrontmatter is returned when the node's frontmatter
// cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrNoOutputs is returned when the target node has no outputs field.
var ErrNoOutputs = errors.New("no outputs")

// ErrPathNotInOutputs is returned when the given path is not declared
// in the node's outputs list.
var ErrPathNotInOutputs = errors.New("path not in outputs")
```

---

## Functions

```go
package mcpwritefile

// MCPWriteFile validates that path is declared in the outputs of the
// node identified by logical_name, then writes content to that path.
// Returns a success message of the form "wrote <path>" on success.
//
// Errors:
//   - ErrUnreadableFrontmatter: the node's frontmatter cannot be parsed.
//   - ErrNoOutputs: the target node has no outputs field.
//   - ErrPathNotInOutputs: path is not declared in the node's outputs.
//   - (LogicalNames.*): propagated from LogicalNameToPath.
//   - (PathUtils.*): propagated from PathValidateCfs.
//   - (FileWriter.*): propagated from FileWrite.
func MCPWriteFile(logical_name string, path string, content string) (string, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
)

func main() {
	result, err := mcpwritefile.MCPWriteFile(
		"ROOT/golang/interfaces/mcp_tools/write_file",
		"code-from-spec/golang/interfaces/mcp_tools/write_file/output.md",
		"# generated content\n",
	)
	if err != nil {
		if errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
			log.Fatal("could not parse node frontmatter")
		}
		if errors.Is(err, mcpwritefile.ErrNoOutputs) {
			log.Fatal("node declares no outputs")
		}
		if errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
			log.Fatal("path is not authorized by node outputs")
		}
		log.Fatalf("write failed: %v", err)
	}

	fmt.Println(result)
}
```
