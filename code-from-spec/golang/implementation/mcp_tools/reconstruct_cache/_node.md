---
depends_on:
  - SPEC/golang/implementation/cache
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpreconstructcache/mcpreconstructcache.go
---

# SPEC/golang/implementation/mcp_tools/reconstruct_cache

Populates the cache from the current state of the
repository. For each manifest entry, resolves the
chain and writes content and chain structure to the
cache. Idempotent — skips files that already exist.

# Public

## Package

`package mcpreconstructcache`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpreconstructcache"`

```go
func MCPReconstructCache() (string, error)
```

### Output

A summary message: `"reconstructed cache: N entries
processed, M content files written, K chain files
written"`.

### Errors

- Propagated errors from `manifest` package.

# Agent

Implement the reconstruct cache tool as a Go package.

## Logic

1. Call `manifest.OpenManifest(true)`. If it fails,
   propagate the error. Store as `m`.

2. Let `entriesProcessed` = 0,
   `contentWritten` = 0,
   `chainWritten` = 0.

3. For each entry in `m.Entries` (iterate in any order):

   a. Derive the spec logical name: strip "ARTIFACT/"
      prefix from the entry key and prepend "SPEC/".

   b. Call `parsing.ParseNode(specName)`. If it fails,
      skip this entry (node may have been deleted).

   c. If `node.Frontmatter.Output` is nil, skip.

   d. Call `chainresolver.ChainResolve(specName)`. If
      it fails, skip this entry.

   e. Call `chainhash.ChainHashCompute(chain)`. It
      returns `(chainHash, positions, err)`. If it
      fails, skip this entry.

   f. For each position in `positions`:
      Extract the content for this position using the
      same logic as load_chain content extraction:
      - If label starts with "AGENT[": extract the
        inner logical name, call
        `parsing.ParseNode(innerName)`, extract agent
        content (same as load_chain instructions
        extraction).
      - If label starts with "INPUT[": extract the
        inner reference name. If it starts with
        "ARTIFACT/" or "EXTERNAL/", read the full file.
        If it starts with "SPEC/", parse the node and
        extract public content (with qualifier if
        present).
      - If label starts with "ARTIFACT/" or
        "EXTERNAL/": read the full file at the
        reference's path.
      - If label starts with "SPEC/": parse the node
        and extract public content (with qualifier if
        present in the label).
      Call `cache.WriteContent(position.Hash, content)`.
      If WriteContent returns nil and the file was
      newly written (not skipped), increment
      `contentWritten`. If it returns an error, skip
      this content file (cache is best-effort).

   g. Call `cache.WriteChain(chainHash, positions)`.
      If WriteChain returns nil and the file was newly
      written, increment `chainWritten`. If it returns
      an error, skip (cache is best-effort).

   h. Increment `entriesProcessed`.

4. Return the summary message with the counts.

## Content extraction

The content extraction for each position must produce
exactly the same text that is hashed by
`ChainHashCompute` and delivered in the spec chain.
Use the same block extraction and concatenation logic
as defined in SPEC/golang/implementation/chain/hash
(ExtractBlock, FormatSection,
ConcatenateSubsections).

For agent content: use the same logic as load_chain
instructions extraction (content of `# Agent`
excluding the heading, with subsections).

For file content (ARTIFACT/ and EXTERNAL/): read the
file using `oslayer.OpenFile` in "read" mode, read
all lines, join with "\n", append trailing "\n".

For SPEC/ content: call `parsing.ParseNode`, then
extract public subsections using
ConcatenateSubsections. For qualified references,
extract only the matching subsection.

## Go-specific guidance

- The package name is `mcpreconstructcache`.
- Use the `manifest` package for `OpenManifest`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`
  and `ContentHash`.
- Use the `cache` package for `WriteContent` and
  `WriteChain`.
- Use the `parsing` package for `ParseNode`.
- Use the `oslayer` package for `OpenFile`, `CfsPath`.
- Use `fmt.Sprintf` for the summary message.
- The block extraction helpers
  (`ExtractBlock`, `FormatSection`,
  `ConcatenateSubsections`) are internal to the
  `chainhash` package. To reuse them, either export
  them from `chainhash` or duplicate the logic. The
  simpler approach is to export them from `chainhash`.
- Cache write errors are silently ignored — log
  nothing, just continue.
