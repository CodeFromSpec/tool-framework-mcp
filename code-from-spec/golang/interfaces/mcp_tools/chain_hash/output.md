[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/chain_hash@09bysZB-Vk9yWMVi27IAO_0aXQI)

# Package `mcpchainhash`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"
```

## Error Sentinels

```go
package mcpchainhash

import "errors"

var ErrNoOutput = errors.New("no output")
```

## Functions

```go
package mcpchainhash

// MCPChainHash resolves the chain for logical_name, computes the 27-character
// base64url chain hash, and returns it.
//
// Returns ErrNoOutput when the target node has no output field.
// Errors from LogicalNameToPath, ChainResolve, ChainHashCompute,
// FrontmatterParse, and FileOpen are propagated as-is.
func MCPChainHash(logical_name string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"
)

func main() {
	hash, err := mcpchainhash.MCPChainHash("ROOT/golang/interfaces/mcp_tools/chain_hash")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hash:", hash)
}
```
