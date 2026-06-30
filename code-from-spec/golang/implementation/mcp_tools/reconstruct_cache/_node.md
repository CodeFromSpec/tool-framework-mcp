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
      Extract the content for this position:
      - If label starts with "AGENT[": extract the
        inner logical name (between "[" and "]"), call
        `parsing.ParseNode(innerName)`, then call
        `parsing.ExtractAgentContent(node)`.
      - If label starts with "INPUT[": extract the
        inner reference name (between "[" and "]").
        Resolve the reference (see below).
      - Otherwise (SPEC/, ARTIFACT/, EXTERNAL/):
        resolve the reference directly from the label.

      Resolving a reference from a label:
      - If it starts with "ARTIFACT/" or "EXTERNAL/":
        call `parsing.CfsReferenceFromName(label)` to
        get the path, then call
        `parsing.ReadFileContent(oslayer.CfsPath(ref.Path))`.
      - If it starts with "SPEC/": parse the qualifier
        if present (text between "(" and ")" at end).
        Call `parsing.ParseNode(logicalName)`. Without
        qualifier: call
        `parsing.ConcatenateSubsections(node.Public.Subsections)`.
        With qualifier: find the matching subsection
        (using `parsing.NormalizeText` for comparison)
        and call `parsing.FormatSection(sub.RawHeading,
        sub.Content)`.

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
Use the extraction helpers from the `parsing` package:
`ConcatenateSubsections`, `FormatSection`,
`ExtractAgentContent`, and `ReadFileContent`.

## Go-specific guidance

- The package name is `mcpreconstructcache`.
- Use the `manifest` package for `OpenManifest`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`
  and `ContentHash`.
- Use the `cache` package for `WriteContent` and
  `WriteChain`.
- Use the `parsing` package for `ParseNode`,
  `CfsReferenceFromName`, `ConcatenateSubsections`,
  `FormatSection`, `ExtractAgentContent`,
  `ReadFileContent`, and `NormalizeText`.
- Use the `oslayer` package for `CfsPath`.
- Use `fmt.Sprintf` for the summary message.
- Cache write errors are silently ignored — log
  nothing, just continue.
