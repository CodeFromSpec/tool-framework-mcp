---
depends_on:
  - SPEC/functional/logic/utils/logical_names
  - SPEC/functional/logic/parsing/frontmatter
  - SPEC/functional/logic/os/path_utils
  - SPEC/functional/logic/os/file
output: code-from-spec/functional/logic/mcp_tools/write_file/output.md
---

# SPEC/functional/logic/mcp_tools/write_file

Writes a generated source file to disk after validating
the path against the node's declared output.

# Public

## Interface

```
function MCPWriteFile(logical_name: string, path: string, content: string) -> string
  errors:
    - QualifierNotAllowed: the logical name contains
      a parenthetical qualifier.
    - UnreadableFrontmatter: the node's frontmatter
      cannot be parsed.
    - NoOutput: target node has no output field.
    - PathNotInOutput: path is not declared in the
      node's output.
    - (LogicalNames.*): propagated from
      LogicalNameToPath.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileReader.*): propagated from FileOpen, FileWrite.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the node whose output authorizes the write. |
| `path` | yes | Relative file path from project root (forward slashes). |
| `content` | yes | Complete file content (UTF-8 text). |

### Output

A success message: `"wrote <path>"`.

# Agent

## Behavior

### Step 1 — Validate and read frontmatter

Check if the logical name has a qualifier using
`LogicalNameHasQualifier`. If true, raise
"qualifier not allowed" — qualifiers reference
subsections and are not valid targets for file
writing.

Resolve the logical name to a file path using
`LogicalNameToPath`. If it fails, propagate the error.
This rejects non-SPEC/ references.

Call `FrontmatterParse` with the resolved path. If
parsing fails, raise "unreadable frontmatter".

If `frontmatter.output` is empty, raise "no output".

### Step 2 — Validate path

Construct a `PathCfs` from the `path` string. Call
`PathValidateCfs` on the path string. If it fails,
propagate the error.

### Step 3 — Check path against output

Compare `path` against `frontmatter.output` (exact
string match). If no match, raise "path not in output".

### Step 4 — Write file

Call `FileOpen` with the `PathCfs`, mode `"overwrite"`, and timeout 30000.
If it fails, propagate the error.

Call `FileWrite` with the handle and `content`. If it
fails, call `FileClose` and propagate the error.

Call `FileClose` to release the handle.

Return `"wrote <path>"` where `<path>` is the path
string.

## Contracts

- Only writes to paths declared in the node's `output`.
- Path validation uses `PathValidateCfs` — same rules
  as the OS layer.
- File writing uses `FileOpen`/`FileWrite`/`FileClose` —
  creates intermediate directories, overwrites existing
  files, acquires exclusive lock during write.
- Writes exactly one file per call.

# Private

## Transport constraints

All MCP tool parameters are transmitted as JSON. The
`content` parameter is a JSON string, which means it is
UTF-8 text. Binary files cannot be written through this
tool. This is consistent with the framework operating
exclusively on text files (see CODE_FROM_SPEC.md).
