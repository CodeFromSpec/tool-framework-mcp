[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/write_file@qf9I293Z7WU0k4e4UgnqLCVybDg)

# Package `mcpwritefile`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpwritefile"
```

Package `mcpwritefile` implements the MCP `write_file` tool. It validates that the requested output path is declared in the target node's frontmatter before delegating the write to the file writer.

---

## Error Sentinels

```go
package mcpwritefile

import "errors"

// ErrUnreadableFrontmatter is returned when the node's frontmatter
// cannot be parsed.
var ErrUnreadableFrontmatter = errors.New("unreadable frontmatter")

// ErrNoOutputs is returned when the target node has no outputs field
// declared in its frontmatter.
var ErrNoOutputs = errors.New("node has no outputs")

// ErrPathNotInOutputs is returned when the requested path is not
// declared in the node's outputs list.
var ErrPathNotInOutputs = errors.New("path not declared in node outputs")
```

---

## Functions

```go
package mcpwritefile

// MCPWriteFile writes content to the given path, provided that path is
// declared in the outputs field of the node identified by logical_name.
//
// The function resolves logical_name to the node's spec file path,
// parses its frontmatter, checks that path appears in the outputs list,
// validates the path, and finally writes the content.
//
// On success it returns the string "wrote <path>".
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
	logicalName := "ROOT/golang/interfaces/mcp_tools/write_file"
	outputPath := "code-from-spec/golang/interfaces/mcp_tools/write_file/output.md"
	content := "# Generated output\n\nHello, world!\n"

	result, err := mcpwritefile.MCPWriteFile(logicalName, outputPath, content)
	if err != nil {
		if errors.Is(err, mcpwritefile.ErrUnreadableFrontmatter) {
			log.Fatal("could not parse node frontmatter")
		}
		if errors.Is(err, mcpwritefile.ErrNoOutputs) {
			log.Fatal("node declares no outputs")
		}
		if errors.Is(err, mcpwritefile.ErrPathNotInOutputs) {
			log.Fatal("path is not authorised by the node's outputs list")
		}
		log.Fatalf("unexpected error: %v", err)
	}

	fmt.Println(result) // wrote code-from-spec/golang/interfaces/mcp_tools/write_file/output.md
}
```
