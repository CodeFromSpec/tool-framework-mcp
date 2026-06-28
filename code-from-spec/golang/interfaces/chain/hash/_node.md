---
depends_on:
  - ARTIFACT/golang/interfaces/chain/resolver
output: code-from-spec/golang/interfaces/chain/hash/output.md
---

# SPEC/golang/interfaces/chain/hash

Computes the chain hash for a resolved chain by reading
all chain positions from disk and hashing their content.

# Public

## Package

`package chainhash`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/chainhash"`

## Interface

```go
func ChainHashCompute(chain *chainresolver.Chain) (string, error)
```

Receives a `Chain` (as returned by `ChainResolve`) and
returns a 27-character base64url encoded SHA-1 hash.

### Errors

- `ErrParseFailure`: a node file cannot be parsed.
- Propagated errors from `file`, `parsenode` packages.

# Agent

Generate an interface specification document listing
the package, import path, and function signatures.
