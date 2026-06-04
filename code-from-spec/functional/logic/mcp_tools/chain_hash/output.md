<!-- code-from-spec: ROOT/functional/logic/mcp_tools/chain_hash@-NhQ1arOBLGCKMGTZSsFLIDg-2c -->

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

MCPChainHash validates the target node, resolves the
chain, and returns the 27-character base64url chain hash.

### Step 1 — Validate

1. Call LogicalNameToPath(logical_name).
   If it raises an error, propagate it.

2. Call FrontmatterParse with the resolved file path.
   If it raises an error, propagate it.

3. If frontmatter.output is empty, raise error NoOutput.

### Step 2 — Resolve chain

4. Call ChainResolve(logical_name).
   If it raises an error, propagate it.

### Step 3 — Compute hash

5. Call ChainHashCompute(chain) with the resolved Chain.
   If it raises an error, propagate it.

6. Return the resulting 27-character hash string.
```
