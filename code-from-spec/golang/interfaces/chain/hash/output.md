[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@VjjWAInRCFZfBMBvYx2JPwJsNV8)

# Package `chainhash`

**Import path:** `github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash`

---

## Error Sentinels

```go
package chainhash

import "errors"

var ErrParseFailure = errors.New("a node file cannot be parsed")
```

---

## Functions

```go
package chainhash

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"

// ChainHashCompute receives a Chain (as returned by ChainResolve) and returns
// a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a content
// hash (SHA-1) for each, concatenates all content hashes as raw bytes in
// chain assembly order, and computes the final SHA-1 of the concatenation.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

---

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/invoices")
	if err != nil {
		log.Fatal(err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chain hash:", hash)
}
```
