---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/path_utils
  - ROOT/functional/logic/os/file_writer
output: code-from-spec/functional/logic/mcp_tools/write_file/output.md
---

# ROOT/functional/logic/mcp_tools/write_file

Writes a generated source file to disk after validating
the path against the node's declared output.

# Public

## Interface

```
function MCPWriteFile(logical_name: string, path: string, content: string) -> string
  errors:
    - UnreadableFrontmatter: the node's frontmatter
      cannot be parsed.
    - NoOutput: target node has no output field.
    - PathNotInOutput: path is not declared in the
      node's output.
    - (LogicalNames.*): propagated from
      LogicalNameToPath.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileWriter.*): propagated from FileWrite.
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

### Step 1 — Read frontmatter

Resolve the logical name to a file path using
`LogicalNameToPath`. If it fails, propagate the error.
This rejects non-ROOT/ references and qualifiers.

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

Call `FileWrite` with the `PathCfs` and `content`. If
it fails, propagate the error.

Return `"wrote <path>"` where `<path>` is the path
string.

## Contracts

- Only writes to paths declared in the node's `output`.
- Path validation uses `PathValidateCfs` — same rules
  as the OS layer.
- File writing uses `FileWrite` — creates intermediate
  directories, overwrites existing files.
- Writes exactly one file per call.

# Private

## Transport constraints

All MCP tool parameters are transmitted as JSON. The
`content` parameter is a JSON string, which means it is
UTF-8 text. Binary files cannot be written through this
tool. This is consistent with the framework operating
exclusively on text files (see CODE_FROM_SPEC.md).
