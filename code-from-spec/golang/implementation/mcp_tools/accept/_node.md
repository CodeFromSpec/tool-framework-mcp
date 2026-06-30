---
depends_on:
  - SPEC/golang/implementation/manifest
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
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
- Propagated errors from `parsing`, `manifest`,
  `oslayer` packages.

# Agent

Implement the accept tool as a Go package.

## Logic

1. If logical_name does not start with "SPEC/",
   return ErrNotASpecReference.

2. Call `parsing.ParseNode(logical_name)`.
   If it fails, return ErrUnreadableFrontmatter.
   Store as node.

4. If `node.Frontmatter.Output` is nil, return ErrNoOutput.

5. Derive the artifact logical name: strip "SPEC/"
   prefix from logical_name and prepend "ARTIFACT/".

6. Call `manifest.OpenManifest(false)`. If it fails,
   propagate the error. Store as m.

7. Look up the artifact logical name in
   m.Entries. If no entry exists,
   call `m.Discard()` and
   return ErrNotModified.

8. Construct oslayer.CfsPath from `*node.Frontmatter.Output`.
   Call `oslayer.OpenFile(path, "read", 30000)`. If it fails,
   call `m.Discard()` and
   propagate the error.

9. Read the full file content. Compute its SHA-1
   hash (base64url, 27 chars) using the same
   normalization as write_file (CRLF→LF, trailing
   LF). Call `handle.Close()`.

10. If the computed hash equals
    entry.Checksum, call
    `m.Discard()` and return
    ErrNotModified (file matches manifest).

11. Update entry.Checksum to the computed hash.

12. Call `m.Save()`. If it
    fails, propagate the error.

13. Return "accepted <*node.Frontmatter.Output>".

## Go-specific guidance

- Use the `parsing` package for `ParseNode`.
- Use the `manifest` package for `OpenManifest`,
  `Manifest`, `ManifestEntry`.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, and `CfsPath`.
- Use `crypto/sha1` and `encoding/base64`
  (base64.RawURLEncoding) for checksum computation.
- Define sentinel errors: `ErrNotASpecReference`,
  `ErrUnreadableFrontmatter`, `ErrNoOutput`,
  `ErrNotModified`.
- The package name should be `mcpaccept`.
