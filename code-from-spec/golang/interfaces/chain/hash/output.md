# code-from-spec: SPEC/golang/interfaces/chain/hash@6Qlm7Pu3jfHYDTuVvLTOHoOV9FE

## Package

```go
package chainhash
```

## Import Path

```
github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash
```

## Error Sentinels

```go
package chainhash

import "errors"

var ErrParseFailure = errors.New("parse failure")
```

## Functions

```go
package chainhash

import (
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

// ChainHashCompute receives a Chain (as returned by ChainResolve) and
// returns a 27-character base64url encoded SHA-1 hash.
//
// The function reads each position's content from disk, computes a content
// hash (SHA-1) for each, concatenates all content hashes as raw bytes in
// chain assembly order, and computes the final SHA-1 of the concatenation.
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"
	"github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainresolver"
)

func main() {
	chain, err := chainresolver.ChainResolve("SPEC/payments/fees")
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
