---
depends_on:
  - ROOT/functional/logic/chain/resolver
  - ROOT/functional/logic/chain/hash
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/path_utils(interface)
  - ROOT/functional/logic/utils/logical_names(interface)
outputs:
  - id: chain_hash
    path: code-from-spec/functional/logic/mcp_tools/chain_hash/output.md
---

# ROOT/functional/logic/mcp_tools/chain_hash

Computes the chain hash for a given node without
assembling the full context stream. Lighter than
`load_chain` when only the hash is needed.

# Public

## Namespace

    namespace: mcpchainhash

## Interface

```
function MCPChainHash(logical_name: string) -> string
  errors:
    - NoOutput: target node has no output field.
    - (LogicalNames.*): propagated from
      LogicalNameToPath.
    - (ChainResolver.*): propagated from ChainResolve.
    - (ChainHash.*): propagated from ChainHashCompute.
    - (Frontmatter.*): propagated from FrontmatterParse.
    - (FileReader.*): propagated from FileOpen.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the target node. |

### Output

The 27-character base64url chain hash.

# Agent

## Behavior

### Step 1 — Validate

Resolve the logical name to a file path using
`LogicalNameToPath`. If it fails, propagate the error.

Read the target node's frontmatter using
`FrontmatterParse`. If `frontmatter.outputs` is empty,
raise NoOutput.

### Step 2 — Resolve chain

Call `ChainResolve(logical_name)` to get the resolved
`Chain`. If it fails, propagate the error.

### Step 3 — Compute hash

Call `ChainHashCompute(chain)` with the resolved Chain.
If it fails, propagate the error. Return the result.

## Contracts

- Returns only the hash string — no context, no input.
- If any file in the chain is unreadable, returns an
  error.
