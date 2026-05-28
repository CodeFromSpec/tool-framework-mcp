---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/os/path_utils
outputs:
  - id: write_file
    path: code-from-spec/functional/logic/mcp_tools/write_file/output.md
---

# ROOT/functional/logic/mcp_tools/write_file

Writes a generated source file to disk after validating the
path against the node's declared outputs.

Review status: pending

# Public

## Interface

```
function WriteFile(logical_name: string, path: PathCfs, content: string) -> string
  errors:
    - invalid logical name: not a recognized ROOT/ reference.
    - no outputs: target node has no outputs field.
    - path validation failure: path is empty, absolute, traversal, or escapes root.
    - path not in outputs: path is not declared in the node's outputs.
    - directory creation failure: cannot create intermediate directories.
    - write failure: cannot write the file to disk.
```

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the node whose outputs authorize the write. |
| `path` | yes | Relative file path from project root. |
| `content` | yes | Complete file content as a JSON string (UTF-8 text). |

### Output

A success message: `"wrote <path>"`.

# Agent

## Behavior

### Validation

Before writing:
1. The logical name must be a valid `ROOT/` reference.
2. Read the frontmatter of the node identified by
   `logical_name`. It must have `outputs` declared.
3. The path must be safe:
   a. Not empty.
   b. Not absolute (no leading `/` or drive letter).
   c. No `..` components after normalization.
   d. After resolving symlinks, must remain within the
      project root.
4. The path must appear in the node's `outputs` list (matched
   against the `path` field of each output entry).

### Write behavior

- Creates intermediate directories if they do not exist.
- Overwrites the file if it already exists.
- Writes exactly one file per call.

## Contracts

- Only writes to paths declared in the node's `outputs`.
- Creates intermediate directories as needed.
- Overwrites existing files without warning.

# Private

## Transport constraints

All MCP tool parameters are transmitted as JSON. The
`content` parameter is a JSON string, which means it is
UTF-8 text. Binary files cannot be written through this
tool. This is consistent with the framework operating
exclusively on text files (see CODE_FROM_SPEC.md).
