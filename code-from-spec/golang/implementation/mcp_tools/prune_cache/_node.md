---
depends_on:
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/manifest
output: internal/mcpprunecache/mcpprunecache.go
---

# SPEC/golang/implementation/mcp_tools/prune_cache

Removes unreferenced files from the cache. Chain files
not referenced by any manifest entry are deleted.
Content files not referenced by any remaining chain
file are deleted.

# Public

## Package

`package mcpprunecache`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpprunecache"`

```go
func MCPPruneCache() (string, error)
```

### Output

A summary message: `"pruned cache: N chain files
deleted, M content files deleted"`.

### Errors

- Propagated errors from `manifest`, `cache` packages.

# Agent

Implement the prune cache tool as a Go package.

## Logic

1. Call `manifest.OpenManifest(true)`. If it fails,
   propagate the error. Store as `m`.

2. Collect all chain hashes referenced by the manifest:
   let `referencedChains` = set of `entry.ChainHash`
   for all entries in `m.Entries`.

3. Call `cache.ListChainHashes()`. If it fails,
   propagate the error. Store as `allChains`.

4. Let `chainsDeleted` = 0.
   For each hash in `allChains`:
   If hash is not in `referencedChains`:
     Call `cache.DeleteChain(hash)`. If it fails,
     skip (best-effort). Else increment
     `chainsDeleted`.

5. Collect all content hashes referenced by remaining
   chain files:
   Let `referencedContent` = empty set.
   For each hash in `referencedChains`:
     Call `cache.ReadChain(hash)`. If it returns
     positions (no error):
       For each position, add position.Hash to
       `referencedContent`.
     If it returns `cache.ErrNotFound` or
     `cache.ErrChainFileCorrupted` or any error, skip.

6. Call `cache.ListContentHashes()`. If it fails,
   propagate the error. Store as `allContent`.

7. Let `contentDeleted` = 0.
   For each hash in `allContent`:
   If hash is not in `referencedContent`:
     Call `cache.DeleteContent(hash)`. If it fails,
     skip (best-effort). Else increment
     `contentDeleted`.

8. Return the summary message with the counts.

## Go-specific guidance

- The package name is `mcpprunecache`.
- Use the `manifest` package for `OpenManifest`.
- Use the `cache` package for `ListChainHashes`,
  `ListContentHashes`, `ReadChain`, `DeleteChain`,
  `DeleteContent`.
- Use the `chainhash` package for `ContentHash`.
- Use a `map[string]bool` for the referenced hash
  sets.
- Use `fmt.Sprintf` for the summary message.
