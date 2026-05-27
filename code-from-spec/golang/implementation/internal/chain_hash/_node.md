# ROOT/golang/implementation/internal/chain_hash

Computes the chain hash for a node by reading raw file
content from disk.

# Public

## Package

`package chainhash`

## Interface

```go
func ComputeChainHash(logicalName string) (string, error)
```

Returns the 27-character base64url chain hash, or an error.
Reads all chain positions raw from disk. Only normalizes
CRLF → LF before hashing.
