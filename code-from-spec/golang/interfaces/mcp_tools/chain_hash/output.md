<!-- code-from-spec: ROOT/golang/interfaces/mcp_tools/chain_hash@phfVnJgfHfm1WI7NhVwTjnvJYaw -->

# Package `mcpchainhash`

```go
package mcpchainhash
```

Import path: `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"`

## Error Sentinels

```go
package mcpchainhash

import "errors"

var ErrNoOutput = errors.New("target node has no output field")
```

## Functions

```go
package mcpchainhash

// MCPChainHash computes and returns the 27-character base64url chain hash
// for the given logical name.
//
// Returns ErrNoOutput if the target node has no output field.
// Propagates errors from LogicalNames, ChainResolver, ChainHash,
// Frontmatter, and FileReader packages.
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
	hash, err := mcpchainhash.MCPChainHash("ROOT/golang/interfaces/chain/resolver")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("chain hash:", hash)
}
```
