[//]: # (code-from-spec: SPEC/golang/interfaces/mcp_tools/chain_hash@8J5MbOBgYELJl-ov6hbPbsTnmAA)

## Package

```go
package mcpchainhash
```

## Import Path

```
github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpchainhash
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

// MCPChainHash resolves the chain for the given logical name, computes
// the 27-character base64url SHA-1 chain hash, and returns it.
//
// Returns ErrNoOutput if the target node has no output field.
// Errors from LogicalNameToPath, ChainResolve, ChainHashCompute,
// FrontmatterParse, and FileOpen are propagated unchanged.
func MCPChainHash(logicalName string) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpchainhash"
)

func main() {
	hash, err := mcpchainhash.MCPChainHash("SPEC/payments/fees")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chain hash:", hash)
}
```
