---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
output: internal/mcpwritefile/mcpwritefile.go
---

# SPEC/golang/implementation/mcp_tools/write_file

Writes a generated source file to disk after validating
the path against the node's declared output.

# Public

## Package

`package mcpwritefile`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpwritefile"`

## Interface

```go
func MCPWriteFile(logicalName, path, content string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the node whose output authorizes the write. |
| `path` | yes | Relative file path from project root (forward slashes). |
| `content` | yes | Complete file content (UTF-8 text). |

### Output

A success message: `"wrote <path>"`.

### Errors

- `ErrNotASpecReference`: the logical name is not a
  SPEC/ reference.
- `ErrQualifierNotAllowed`: the logical name contains
  a parenthetical qualifier.
- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no output field.
- `ErrPathNotInOutput`: path is not declared in the
  node's output.
- Propagated errors from `parsing`, `oslayer`
  packages.

# Agent

Implement the write file tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/",
   return ErrNotASpecReference.

2. Call `parsing.ParseNode(logical_name)`.
   If it fails, return ErrUnreadableFrontmatter.
   Store the result as node.

3. If node.Reference.Qualifier is not nil, return error
   ErrQualifierNotAllowed.

5. If `node.Frontmatter.Output` is nil, return error
   ErrNoOutput.

6. Call `oslayer.ValidateStringIsCfsPath` with path. If it fails,
   propagate the error.

7. If path does not exactly match `*node.Frontmatter.Output`,
   return ErrPathNotInOutput.

8. Construct an `oslayer.CfsPath` record with value set to
   path. Call `oslayer.OpenFile` with that CfsPath, mode "overwrite",
   and timeout 30000. If it fails, propagate the error.
   Store the result as handle.

9. Call `handle.Write(content)`. If it fails, call
   `handle.Close()`, then propagate the error.

10. Call `handle.Close()`.

11. Compute the checksum of `content`: SHA-1 of the
    content bytes (after CRLF→LF normalization and
    ensuring a trailing LF), encoded as base64url
    (27 characters).

12. Call `chainresolver.ChainResolve(logical_name)`. If it fails,
    propagate the error.

13. Call `chainhash.ChainHashCompute(chain)`. If it fails,
    propagate the error.

14. Call `manifest.OpenManifest(false)`. If it fails,
    propagate the error.

15. Derive the artifact logical name: strip "SPEC/"
    prefix from logical_name and prepend "ARTIFACT/".
    Set m.Entries[artifact_name] =
    ManifestEntry{Path: path, Checksum: checksum,
    ChainHash: chain_hash}.

16. Call `m.Save()`. If it fails,
    propagate the error.

17. Return "wrote <path>" where <path> is the path
    string.

## Go-specific guidance

- Use the `parsing` package for `ParseNode` and
  `Node`.
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
- The package name should be `mcpwritefile`.
- The function receives plain strings from the MCP
  transport layer. Construct `CfsPath` internally.
