---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpaccept/mcpaccept.go
---

# SPEC/golang/implementation/mcp_tools/accept

Accepts an artifact without regenerating it. Updates
the manifest entry to match the current state:
checksum from the file on disk, chain hash from the
current spec tree.

# Public

## Package

`package mcpaccept`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpaccept"`

```go
func MCPAccept(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the node whose artifact should be accepted. |

### Output

A success message: `"accepted <artifact_path>"`.

### Errors

- `ErrNotASpecReference`: the logical name is not a
  SPEC/ reference.
- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no output field.
- `ErrAlreadyUpToDate`: the artifact is already up to
  date (manifest entry exists and both checksum and
  chain hash match current values).
- Propagated errors from `parsing`, `manifest`,
  `oslayer`, `chainresolver`, `chainhash` packages.

# Agent

Implement the accept tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/",
   return ErrNotASpecReference.

2. Call `parsing.ParseNode(logical_name)`.
   If it fails, return ErrUnreadableFrontmatter.
   Store as node.

3. If `node.Frontmatter.Output` is nil, return ErrNoOutput.

4. Derive the artifact logical name: strip "SPEC/"
   prefix from logical_name and prepend "ARTIFACT/".

5. Construct oslayer.CfsPath from `*node.Frontmatter.Output`.
   Call `oslayer.OpenFile(path, "read", 30000)`. If it
   fails, propagate the error.

6. Read the full file content. Compute its SHA-1
   hash (base64url, 27 chars) using the same
   normalization as write_file (CRLF→LF, trailing
   LF). Call `handle.Close()`. Store as `checksum`.

7. Call `chainresolver.ChainResolve(logical_name)`.
   If it fails, propagate the error.

8. Call `chainhash.ChainHashCompute(chain)`. It returns
   `(chainHash, positions, err)`. If it fails,
   propagate the error. Ignore `positions`.

9. Call `manifest.OpenManifest(false)`. If it fails,
   propagate the error. Store as m.
   Defer `m.Discard()`.

10. Look up the artifact logical name in m.Entries.
    If no entry exists:
      Set m.Entries[artifactName] =
      ManifestEntry{Path: *node.Frontmatter.Output,
      Checksum: checksum, ChainHash: chainHash}.
      Call `m.Save()`. Return
      "accepted <*node.Frontmatter.Output>".

11. If entry exists and entry.Checksum equals checksum
    and entry.ChainHash equals chainHash:
      Return ErrAlreadyUpToDate.

12. Update entry.Checksum to checksum.
    Update entry.ChainHash to chainHash.

13. Call `m.Save()`. If it fails, propagate the error.

14. Return "accepted <*node.Frontmatter.Output>".

## Go-specific guidance

- Use the `parsing` package for `ParseNode`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, and `CfsPath`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- Define sentinel errors: `ErrNotASpecReference`,
  `ErrUnreadableFrontmatter`, `ErrNoOutput`,
  `ErrAlreadyUpToDate`.
- The package name should be `mcpaccept`.
