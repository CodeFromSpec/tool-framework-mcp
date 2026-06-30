---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/prune_cache
  - SPEC/golang/implementation/oslayer(interface)
output: internal/mcpprunecache/mcpprunecache_test.go
---

# SPEC/golang/test/cases/mcp_tools/prune_cache

# Agent

## Test setup guidance

`MCPPruneCache` reads the manifest and cache to
determine which files to delete. Tests must create
cache files and manifest entries to set up the
conditions.

Use `testutils.Chdir` for isolation. Create cache
files with `cache.WriteContent` and
`cache.WriteChain`. Create manifest entries with
`manifest.OpenManifest(false)` + `m.Save()`.

## Test cases

### Happy path

#### Deletes unreferenced chain file

Setup:
- Create a manifest with one entry whose ChainHash is
  "aaaaaaaaaaaaaaaaaaaaaaaaaa1".
- Create chain files:
  - "aaaaaaaaaaaaaaaaaaaaaaaaaa1" (referenced)
  - "bbbbbbbbbbbbbbbbbbbbbbbbbbb" (unreferenced)
  Both with at least one position.
- Create content files for all positions in both chains.

Actions:
1. Call `mcpprunecache.MCPPruneCache()`.

Expected:
- No error.
- Summary contains "1 chain files deleted".
- `cache.ReadChain("bbbbbbbbbbbbbbbbbbbbbbbbbbb")`
  returns `cache.ErrNotFound`.
- `cache.ReadChain("aaaaaaaaaaaaaaaaaaaaaaaaaa1")`
  still succeeds.

#### Deletes unreferenced content file

Setup:
- Create a manifest with one entry whose ChainHash is
  "aaaaaaaaaaaaaaaaaaaaaaaaaa1".
- Create chain file "aaaaaaaaaaaaaaaaaaaaaaaaaa1" with
  positions referencing content hash "ccc...".
- Create content files:
  - "ccc..." (referenced by the chain)
  - "ddd..." (not referenced by any chain)

Actions:
1. Call `mcpprunecache.MCPPruneCache()`.

Expected:
- No error.
- Summary contains "1 content files deleted".
- `cache.ReadContent("ddd...")` returns
  `cache.ErrNotFound`.
- `cache.ReadContent("ccc...")` still succeeds.

#### Nothing to prune — zero deleted

Setup:
- Create a manifest and cache where all files are
  referenced.

Actions:
1. Call `mcpprunecache.MCPPruneCache()`.

Expected:
- No error.
- Summary contains "0 chain files deleted, 0 content
  files deleted".

### Edge cases

#### Empty cache — no errors

Setup:
- Create a manifest with entries but no cache
  directory.

Actions:
1. Call `mcpprunecache.MCPPruneCache()`.

Expected:
- No error. Summary shows 0 deleted.

#### Empty manifest — all cache files pruned

Setup:
- Create a valid manifest with header only (no
  entries). Create the `.manifest.lock` file.
- Create cache files (chain and content).

Actions:
1. Call `mcpprunecache.MCPPruneCache()`.

Expected:
- All chain and content files deleted.

## Go-specific guidance

- The package name is `mcpprunecache_test` (external
  test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- Use `cache.WriteContent` and `cache.WriteChain` to
  create cache fixtures.
- Use `manifest.OpenManifest(false)` + `m.Save()` to
  create manifest fixtures.
- Verify deletions with `cache.ReadContent` and
  `cache.ReadChain`, expecting `cache.ErrNotFound`.
- Use `errors.Is` for error sentinel checks.
