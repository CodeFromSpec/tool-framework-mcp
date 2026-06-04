[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@Wb5Aoo5Gv5B0rF5enMl8RT2D3ak)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

## Error Sentinels

```go
package mcploadchain

import "errors"

var ErrNoOutput = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")
```

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the spec chain for the given logical name and returns
// a formatted string containing the chain hash, context, optional input, and
// optional existing artifact sections.
//
// The returned string has the following format:
//
//	chain_hash: <27-character hash>
//	--- context ---
//	<context content>
//	--- input ---
//	<input content>
//	--- existing artifact ---
//	<existing artifact content>
//
// The "--- input ---" section is only present when the target node's
// frontmatter has a non-empty input field. The "--- existing artifact ---"
// section is only present when the output file exists on disk and is readable.
func MCPLoadChain(logical_name string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
)

func main() {
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```
