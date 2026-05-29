[//]: # (code-from-spec: ROOT/golang/interfaces/chain/hash@gSFxwhOyZl245SHIAvewTpqF1Bk)

# Interface: `chainhash`

**Package:** `package chainhash`  
**Import:** `import "github.com/CodeFromSpec/tool-framework-mcp/v3/internal/chainhash"`

---

## Functions

```go
// ChainHashCompute receives a Chain (as returned by chainresolver.ChainResolve)
// and returns a 27-character base64url-encoded SHA-1 hash.
//
// The function reads each position's content from disk in chain assembly order:
//  1. Ancestors — root down to (but not including) the target.
//  2. Dependencies — target's depends_on, sorted alphabetically by file path
//     then qualifier.
//  3. External — target's external files, sorted alphabetically by path.
//  4. Target — the target node itself.
//  5. Input — the target's input artifact, if present.
//
// For each position, the raw file bytes are read from disk and a SHA-1 content
// hash is computed. All content hashes are concatenated as raw bytes in the
// assembly order above, and a final SHA-1 is computed over the concatenation.
// The result is encoded using base64url (no padding) and truncated to 27
// characters.
//
// Returns an error if:
//   - a file in the chain cannot be read or opened.
//   - a node file cannot be parsed.
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
    chain, err := chainresolver.ChainResolve("ROOT/golang/interfaces/chain/hash")
    if err != nil {
        log.Fatalf("failed to resolve chain: %v", err)
    }

    hash, err := chainhash.ChainHashCompute(chain)
    if err != nil {
        log.Fatalf("failed to compute chain hash: %v", err)
    }

    fmt.Println("Chain hash:", hash) // e.g. "gSFxwhOyZl245SHIAvewTpqF1Bk"
}
```
