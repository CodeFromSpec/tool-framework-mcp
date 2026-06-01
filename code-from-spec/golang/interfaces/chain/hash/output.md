[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@L86EFczDZEpblyV2wAS8jb9Maqs)

# Package `chainhash`

```
import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
```

Package `chainhash` computes a deterministic 27-character hash over an ordered chain of spec node files.

---

## Error Sentinels

```go
package chainhash

import "errors"

// ErrFileUnreadable is returned when a file in the chain cannot be read or opened.
var ErrFileUnreadable = errors.New("file unreadable")

// ErrParseFailure is returned when a node file cannot be parsed.
var ErrParseFailure = errors.New("parse failure")
```

---

## Functions

```go
package chainhash

import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"

// ChainHashCompute receives a Chain (as returned by ChainResolve) and returns
// a 27-character base64url encoded SHA-1 hash.
//
// For each position in the chain (in assembly order), the function reads the
// file content from disk and computes a SHA-1 hash of that content. All
// per-file SHA-1 digests are concatenated as raw bytes in chain assembly order,
// and a final SHA-1 is computed over the concatenation. The result is encoded
// as base64url without padding, producing a 27-character string.
//
// Errors:
//   - ErrFileUnreadable: a file in the chain cannot be read or opened.
//   - ErrParseFailure: a node file cannot be parsed.
//   - (FileReader.*): propagated from FileOpen.
//   - (NodeParsing.*): propagated from NodeParse.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

---

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/hash")
	if err != nil {
		log.Fatalf("chain resolution failed: %v", err)
	}

	hash, err := chainhash.ChainHashCompute(chain)
	if err != nil {
		if errors.Is(err, chainhash.ErrFileUnreadable) {
			log.Fatal("a file in the chain could not be read")
		}
		if errors.Is(err, chainhash.ErrParseFailure) {
			log.Fatal("a node file could not be parsed")
		}
		log.Fatalf("hash computation failed: %v", err)
	}

	fmt.Println("chain hash:", hash)
}
```
