---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/mcpchainhash/mcpchainhash.go
---

# SPEC/golang/implementation/mcp_tools/chain_hash

Computes the chain hash for a given node without
assembling the full context stream. Lighter than
`load_chain` when only the hash is needed.

# Public

## Package

`package mcpchainhash`

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/mcpchainhash"`

## Interface

```go
var ErrNoOutput = errors.New("no output")

func MCPChainHash(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the target node. |

### Output

The 27-character base64url chain hash.

### Errors

- `ErrNoOutput`: target node has no `output` field.
- Propagated from `logicalnames.LogicalNameToPath`.
- Propagated from `frontmatter.FrontmatterParse`.
- Propagated from `chainresolver.ChainResolve`.
- Propagated from `chainhash.ChainHashCompute`.

# Agent

Implement the `mcpchainhash` package.

## Steps

### Step 1 — Validate

Call `logicalnames.LogicalNameParse(logicalName)`.
If it fails, wrap and return the error.
Let `ln` be the result.

Parse the target node's frontmatter using
`frontmatter.FrontmatterParse(PathCfs{Value: ln.Path})`.
If `fm.Output` is empty, return `ErrNoOutput`.

### Step 2 — Resolve chain

Call `chainresolver.ChainResolve(logicalName)`.
If it fails, wrap and return the error.

### Step 3 — Compute hash

Call `chainhash.ChainHashCompute(chain)`.
If it fails, wrap and return the error.
Return the hash string.

## Go-specific guidance

- Import `chainresolver`, `chainhash`, `frontmatter`,
  `logicalnames`, `pathutils` packages.
- Wrap each error with `fmt.Errorf("context: %w", err)`
  to preserve the sentinel chain.

## Contracts

- Returns only the hash string — no context, no input.
- If any file in the chain is unreadable, returns an
  error.
