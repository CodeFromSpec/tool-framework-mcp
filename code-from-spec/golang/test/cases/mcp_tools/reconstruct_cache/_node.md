---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/test/utils/create_spec_node
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/mcp_tools/reconstruct_cache
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpreconstructcache/mcpreconstructcache_test.go
---

# SPEC/golang/test/cases/mcp_tools/reconstruct_cache

# Agent

## Test setup guidance

`MCPReconstructCache` reads the manifest, resolves
chains, and populates the cache. Tests must create a
complete spec tree on disk with valid `_node.md` files
and a `.manifest` file with correct entries.

Use `testutils.Chdir` and create
`code-from-spec/.../_node.md` files. To produce valid
manifest entries, use `chainresolver.ChainResolve` and
`chainhash.ChainHashCompute` to compute the chain hash,
and compute the file checksum.

## Test cases

### Happy path

#### Populates cache for single entry

Setup:
- Create `code-from-spec/root/_node.md` with
  `# SPEC/root`, `# Public` → `## Context` with
  content.
- Create `code-from-spec/root/a/_node.md` with
  `# SPEC/root/a`, frontmatter `output: out/a.go`,
  `# Public` → `## Interface` with content.
- Create `out/a.go` with known content.
- Create a valid `.manifest` entry for ARTIFACT/root/a
  with matching chain hash and checksum.

Actions:
1. Call `mcpreconstructcache.MCPReconstructCache()`.

Expected:
- No error.
- Summary message contains "1 entries processed".
- `cache.ReadChain` succeeds for the chain hash from
  the manifest entry.
- `cache.ReadContent` succeeds for at least one
  content hash from the chain positions.

#### Idempotent — skips existing cache files

Setup:
- Same as above. Call MCPReconstructCache once.

Actions:
1. Call `mcpreconstructcache.MCPReconstructCache()` again.

Expected:
- No error.
- Summary message shows 0 content files written and
  0 chain files written on second call.

#### Skips deleted nodes gracefully

Setup:
- Create a `.manifest` with an entry for
  ARTIFACT/root/deleted, but do not create the
  corresponding `_node.md` file.

Actions:
1. Call `mcpreconstructcache.MCPReconstructCache()`.

Expected:
- No error. The deleted entry is skipped.

### Edge cases

#### No manifest — returns error

Setup:
- Do not create a `.manifest` file or lock file.

Actions:
1. Call `mcpreconstructcache.MCPReconstructCache()`.

Expected:
- Returns error (manifest cannot be opened).

#### Empty manifest — zero entries processed

Setup:
- Create a valid `.manifest` with header only (no
  entries). Create the `.manifest.lock` file.

Actions:
1. Call `mcpreconstructcache.MCPReconstructCache()`.

Expected:
- No error.
- Summary message contains "0 entries processed".

## Go-specific guidance

- The package name is `mcpreconstructcache_test`
  (external test package).
- Use `testutils.Chdir(t)` to create a temp dir and
  set the working directory.
- To create valid manifest entries, compute chain hash
  with `chainhash.ChainHashCompute` and file checksum
  with SHA-1 (base64url, 27 chars).
- Write the manifest file using
  `manifest.OpenManifest(false)` + `m.Save()`.
- Verify cache contents with `cache.ReadChain` and
  `cache.ReadContent`.
