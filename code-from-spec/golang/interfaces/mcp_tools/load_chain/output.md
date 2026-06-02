<!-- code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@5vTisVNWFkG1YQl0-n9QFud4Wwo -->

# Package `mcploadchain`

```go
package mcploadchain
```

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"`

## Struct Definitions

```go
package mcploadchain

// MCPLoadChainResult holds the result of a load_chain tool call.
type MCPLoadChainResult struct {
	ChainHash string
	Context   string
	Input     *string
}
```

## Error Sentinels

```go
package mcploadchain

import "errors"

var ErrNoOutput          = errors.New("target node has no output field")
var ErrInvalidOutputPath = errors.New("the output path fails path validation")
```

## Functions

```go
package mcploadchain

// MCPLoadChain resolves, concatenates, and returns the full chain content
// for the given logical name, along with its 27-character base64url chain
// hash and optional input artifact content.
//
// Returns ErrNoOutput if the target node has no output field.
// Returns ErrInvalidOutputPath if the output path fails path validation.
// Propagates errors from LogicalNames, ChainResolver, ChainHash,
// NodeParsing, and FileReader packages.
func MCPLoadChain(logical_name string) (*MCPLoadChainResult, error)
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
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("chain hash:", result.ChainHash)
	fmt.Println("context length:", len(result.Context))

	if result.Input != nil {
		fmt.Println("input length:", len(*result.Input))
	}
}
```
