---
depends_on:
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/utils/logical_names
output: internal/mcpaccept/mcpaccept.go
---

# SPEC/golang/implementation/mcp_tools/accept

Accepts a modified artifact without regenerating it.
Updates the manifest checksum to match the current
file on disk.

# Public

## Package

`package mcpaccept`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/mcpaccept"`

## Interface

```go
func MCPAccept(logicalName string) (string, error)
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logicalName` | yes | Logical name of the node whose artifact was modified. |

### Output

A success message: `"accepted <artifact_path>"`.

### Errors

- `ErrNotASpecReference`: the logical name is not a
  SPEC/ reference.
- `ErrUnreadableFrontmatter`: the node's frontmatter
  cannot be parsed.
- `ErrNoOutput`: target node has no output field.
- `ErrNotModified`: the artifact is not in modified
  status (checksum in manifest matches file on disk,
  or no manifest entry exists).
- Propagated errors from `logicalnames`, `manifest`,
  `file`, `pathutils` packages.

# Agent

Implement the accept tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/",
   return error "not a SPEC reference".

2. Call `LogicalNameParse(logical_name)`.
   If it fails, propagate the error.
   Let `ln` be the result.

3. Call `FrontmatterParse(PathCfs{Value: ln.Path})`.
   If it fails, return error "unreadable frontmatter".
   Store as frontmatter.

4. If `frontmatter.output` is empty, return error
   "no output".

5. Derive the artifact logical name: strip "SPEC/"
   prefix from logical_name and prepend "ARTIFACT/".

6. Call `ManifestOpen("write")`. If it fails,
   propagate the error. Store as manifest_handle.

7. Look up the artifact logical name in
   manifest_handle.Entries. If no entry exists,
   call `ManifestDiscard(manifest_handle)` and
   return error "not modified".

8. Construct PathCfs from frontmatter.output. Call
   `FileOpen(path, "read", 30000)`. If it fails,
   call `ManifestDiscard(manifest_handle)` and
   propagate the error.

9. Read the full file content. Compute its SHA-1
   hash (base64url, 27 chars) using the same
   normalization as write_file (CRLF→LF, trailing
   LF). Call `FileClose`.

10. If the computed hash equals
    entry.Checksum, call
    `ManifestDiscard(manifest_handle)` and return
    error "not modified" (file matches manifest).

11. Update entry.Checksum to the computed hash.

12. Call `ManifestSave(manifest_handle)`. If it
    fails, propagate the error.

13. Return "accepted <frontmatter.output>".

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameParse`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `manifest` package for `ManifestOpen`,
  `ManifestSave`, `ManifestDiscard`, `ManifestEntry`.
- Use the `file` package for `FileOpen`, `FileReadLine`,
  `FileClose`.
- Use the `pathutils` package for `PathCfs`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- Define sentinel errors: `ErrNotASpecReference`,
  `ErrUnreadableFrontmatter`, `ErrNoOutput`,
  `ErrNotModified`.
- The package name should be `mcpaccept`.
