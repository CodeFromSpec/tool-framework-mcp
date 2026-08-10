---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/subagent_token(interface)
output: internal/mcpwriteverdict/mcpwriteverdict.go
---

# SPEC/golang/implementation/mcp_tools/write_verdict

Writes a verdict document to disk and updates the
manifest. The output path is derived from the node's
frontmatter — the caller provides an opaque token
identifying the node (see `mcp_tools/create_token`),
the pass/fail result, and the document content.

# Public

## Package

`package mcpwriteverdict`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/mcpwriteverdict"`

## Interface

```go
func MCPWriteVerdict(token string, passed bool, content string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `token` | yes | Opaque token identifying the node whose output declares the target path, as returned by `create_token`. |
| `passed` | yes | The verdict: true for pass, false for fail. |
| `content` | yes | Complete verdict document content (UTF-8 text). |

### Output

A success message: `"wrote <path>"`, where `<path>` is
the output path read from the node's frontmatter.

### Errors

- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no type field.
- `ErrNotAVerdict`: target node's type is not
  `"verdict"`.
- Propagated errors from `subagenttoken`, `parsing`,
  `oslayer` packages.

# Agent

Implement the write verdict tool as a Go package.

## Logic

1. Call `subagenttoken.SubagentTokenValidate(token)` to
   recover the target node's logical name. If it fails,
   propagate the error. Store the result as
   `logical_name`.

2. Call `parsing.ParseNode(logical_name)`.
   If it fails, return ErrUnreadableFrontmatter.
   Store the result as node.

3. If `node.Frontmatter.Type` is nil, return ErrNoOutput.
   If `*node.Frontmatter.Type` is not `"verdict"`,
   return ErrNotAVerdict.

4. Let `resolved_output` =
   `parsing.ResolvedOutput(node)`. If `resolved_output`
   is nil, return error ErrNoOutput.

5. Store `*resolved_output` as path.

6. Call `oslayer.ValidateStringIsCfsPath` with path.
   If it fails, propagate the error.

7. Construct an `oslayer.CfsPath` record with value set to
   path. Call `oslayer.OpenFile` with that CfsPath, mode
   "overwrite", and timeout 30000. If it fails,
   propagate the error. Store the result as handle.

8. Call `handle.Write(content)`. If it fails, call
   `handle.Close()`, then propagate the error.

9. Call `handle.Close()`.

10. Compute the checksum of `content`: SHA-1 of the
    content bytes (after CRLF→LF normalization and
    ensuring a trailing LF), encoded as base64url
    (27 characters).

11. Call `chainresolver.ChainResolve(logical_name)`. If
    it fails, propagate the error.

12. Call `chainhash.ChainHashCompute(chain)`. It returns
    `(chain_hash, positions, err)`. If it fails,
    propagate the error.

13. Call `manifest.OpenManifest(false)`. If it fails,
    propagate the error.

14. Derive the verdict logical name: strip "SPEC/"
    prefix from logical_name and prepend "VERDICT/".
    Let result_value = "pass" if passed is true,
    "fail" otherwise.
    Set m.Entries[verdict_name] =
    ManifestEntry{Path: path, Checksum: checksum,
    ChainHash: chain_hash, Result: result_value}.

15. Call `m.Save()`. If it fails,
    propagate the error.

16. Return "wrote <path>" where <path> is the path
    string.

## Go-specific guidance

- Use the `subagenttoken` package for
  `SubagentTokenValidate`.
- Use the `parsing` package for `ParseNode`,
  `ResolvedOutput`, and `Node`.
- Use the `oslayer` package for `ValidateStringIsCfsPath`,
  `CfsPath`, `OpenFile`, `.Write()`, and `.Close()`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- The CRLF→LF normalization and trailing LF for
  checksum must match the normalization used by
  `ChainHashCompute` for whole-file content.
- The package name should be `mcpwriteverdict`.
- The function receives plain strings from the MCP
  transport layer. Construct `CfsPath` internally.
- Do NOT write to cache — verdict chains are never
  cached.
