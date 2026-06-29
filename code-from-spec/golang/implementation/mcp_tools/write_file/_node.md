---
depends_on:
  - SPEC/golang/implementation/chain/hash
  - SPEC/golang/implementation/chain/resolver
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
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
- Propagated errors from `logicalnames`, `oslayer`
  packages.

# Agent

Implement the write file tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/",
   return ErrNotASpecReference.

2. Call `LogicalNameParse(logical_name)`.
   If it fails, propagate the error.
   Let `ln` be the result.

3. If ln.Qualifier is not nil, return error
   ErrQualifierNotAllowed.

4. Call `FrontmatterParse(CfsPath(ln.Path))`.
   If it fails, return ErrUnreadableFrontmatter.
   Store the result as frontmatter.

5. If `frontmatter.output` is empty, return error
   ErrNoOutput.

6. Call `ValidateCfsPath` with path. If it fails,
   propagate the error.

7. If path does not exactly match `frontmatter.output`,
   return ErrPathNotInOutput.

8. Construct a `CfsPath` record with value set to path.
   Call `OpenFile` with that CfsPath, mode "overwrite",
   and timeout 30000. If it fails, propagate the error.
   Store the result as handle.

9. Call `handle.Write(content)`. If it fails, call
   `handle.Close()`, then propagate the error.

10. Call `handle.Close()`.

11. Compute the checksum of `content`: SHA-1 of the
    content bytes (after CRLF→LF normalization and
    ensuring a trailing LF), encoded as base64url
    (27 characters).

12. Call `ChainResolve(logical_name)`. If it fails,
    propagate the error.

13. Call `ChainHashCompute(chain)`. If it fails,
    propagate the error.

14. Call `ManifestOpen("write")`. If it fails,
    propagate the error.

15. Derive the artifact logical name: strip "SPEC/"
    prefix from logical_name and prepend "ARTIFACT/".
    Set manifest_handle.Entries[artifact_name] =
    ManifestEntry{Path: path, Checksum: checksum,
    ChainHash: chain_hash}.

16. Call `ManifestSave(manifest_handle)`. If it fails,
    propagate the error.

17. Return "wrote <path>" where <path> is the path
    string.

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameParse`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `oslayer` package for `ValidateCfsPath`,
  `CfsPath`, `OpenFile`, `.Write()`, and `.Close()`.
- Use the `chainresolver` package for `ChainResolve`.
- Use the `chainhash` package for `ChainHashCompute`.
- Use the `manifest` package for `ManifestOpen`,
  `ManifestSave`, `ManifestEntry`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- The CRLF→LF normalization and trailing LF for
  checksum must match the normalization used by
  `ChainHashCompute` for whole-file content.
- The package name should be `mcpwritefile`.
- The function receives plain strings from the MCP
  transport layer. Construct `CfsPath` internally.
