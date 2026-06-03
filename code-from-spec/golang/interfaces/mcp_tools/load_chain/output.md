[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/load_chain@350bZfPnmiBNaqdl691iII5gdEM)

# Package `mcploadchain`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcploadchain"
```

## Structs

```go
package mcploadchain

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

var ErrNoOutput = errors.New("no output")
var ErrInvalidOutputPath = errors.New("invalid output path")
```

## Functions

```go
package mcploadchain

// MCPLoadChain resolves the spec chain for the given logical name, computes
// its hash, and returns the concatenated chain context along with an optional
// input artifact body.
//
// ErrNoOutput is returned when the target node has no output field.
// ErrInvalidOutputPath is returned when the output path fails path validation.
// Errors from LogicalNameToPath, ChainResolve, ChainHashCompute, NodeParse,
// and FileOpen are propagated directly.
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
	result, err := mcploadchain.MCPLoadChain("ROOT/golang/interfaces/mcp_tools/load_chain")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hash:", result.ChainHash)
	fmt.Println("Context:", result.Context)

	if result.Input != nil {
		fmt.Println("Input:", *result.Input)
	}
}
```
