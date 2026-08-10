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

Accepts an artifact or verdict without regenerating it.
Updates the manifest entry to match the current state:
checksum from the file on disk, chain hash from the
current spec tree. For verdicts, sets the result to
`accepted`.

# Public

## Package

`package mcpaccept`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpaccept"`

```go
func MCPAccept(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the artifact or verdict to accept: `ARTIFACT/<name>` or `VERDICT/<name>`. |

### Output

A success message: `"accepted <output_path>"`.

### Errors

- `ErrInvalidPrefix`: the logical name does not start
  with `ARTIFACT/` or `VERDICT/`.
- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no type field.
- `ErrAlreadyUpToDate`: the entry is already up to
  date (manifest entry exists, both checksum and chain
  hash match current values, and — for `VERDICT/`
  entries — result is already `accepted`).
- Propagated errors from `parsing`, `manifest`,
  `oslayer`, `chainresolver`, `chainhash` packages.

# Agent

Implement the accept tool as a Go package.

## Logic

1. Determine the prefix and derive the spec name:
   If logical_name starts with "ARTIFACT/":
     Let `spec_name` = "SPEC/" + logical_name with
     "ARTIFACT/" prefix removed.
     Let `is_verdict` = false.
   Else if logical_name starts with "VERDICT/":
     Let `spec_name` = "SPEC/" + logical_name with
     "VERDICT/" prefix removed.
     Let `is_verdict` = true.
   Else:
     Return ErrInvalidPrefix.

2. Call `parsing.ParseNode(spec_name)`.
   If it fails, return ErrUnreadableFrontmatter.
   Store as node.

3. Let `resolved_output` =
   `parsing.ResolvedOutput(node)`. If `resolved_output`
   is nil, return ErrNoOutput.

4. Construct oslayer.CfsPath from `*resolved_output`.
   Call `oslayer.OpenFile(path, "read", 30000)`. If it
   fails, propagate the error.

5. Read the full file content. Compute its SHA-1
   hash (base64url, 27 chars) using the same
   normalization as write_file (CRLF→LF, trailing
   LF). Call `handle.Close()`. Store as `checksum`.

6. Call `chainresolver.ChainResolve(spec_name)`.
   If it fails, propagate the error.

7. Call `chainhash.ChainHashCompute(chain)`. It returns
   `(chainHash, positions, err)`. If it fails,
   propagate the error. Ignore `positions`.

8. Call `manifest.OpenManifest(false)`. If it fails,
   propagate the error. Store as m.
   Defer `m.Discard()`.

9. Look up logical_name in m.Entries.
   If no entry exists:
     Let entry = ManifestEntry{Path: *resolved_output,
     Checksum: checksum, ChainHash: chainHash}.
     If is_verdict: set entry.Result = "accepted".
     Set m.Entries[logical_name] = entry.
     Call `m.Save()`. Return
     "accepted <*resolved_output>".

10. If entry exists and entry.Checksum equals checksum
    and entry.ChainHash equals chainHash:
      If is_verdict and entry.Result is not "accepted":
        (continue to step 11 — result needs updating)
      Else:
        Return ErrAlreadyUpToDate.

11. Update entry.Checksum to checksum.
    Update entry.ChainHash to chainHash.
    If is_verdict: set entry.Result = "accepted".

12. Call `m.Save()`. If it fails, propagate the error.

13. Return "accepted <*resolved_output>".

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
- Define sentinel errors: `ErrInvalidPrefix`,
  `ErrUnreadableFrontmatter`, `ErrNoOutput`,
  `ErrAlreadyUpToDate`.
- The package name should be `mcpaccept`.
