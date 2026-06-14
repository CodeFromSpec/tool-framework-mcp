[//]: # (code-from-spec: ROOT/golang/interfaces/mcp_tools/chain_hash@OPaVbgqB-UjCmQD_OtbZsKB0u8A)

# Package `mcpchainhash`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash`

---

## Functions

```go
package mcpchainhash

// MCPChainHash resolves the chain for the given logical name, computes its
// 27-character base64url SHA-1 hash, and returns it.
func MCPChainHash(logical_name string) (string, error)
```

---

## Error Sentinels

```go
package mcpchainhash

import "errors"

var ErrNoOutput = errors.New("target node has no output field")
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/mcpchainhash"
)

func main() {
	hash, err := mcpchainhash.MCPChainHash("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chain hash:", hash)
}
```
