---
outputs:
  - id: write_file
    path: code-from-spec/functional/mcp_tools/write_file/output.md
---

# ROOT/functional/mcp_tools/write_file

Writes a generated source file to disk after validating the
path against the node's declared outputs.

# Public

## Behavior

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the node whose outputs authorize the write. |
| `path` | yes | Relative file path from project root. |
| `content` | yes | Complete file content as a JSON string (UTF-8 text). |

### Output

A success message: `"wrote <path>"`.

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

## Error conditions

| Condition | Description |
|---|---|
| Invalid logical name | Not a recognized `ROOT/` reference. |
| No outputs | Target node has no `outputs` field. |
| Path validation failure | Path is empty, absolute, traversal, or escapes root. |
| Path not in outputs | Path is not declared in the node's `outputs`. |
| Directory creation failure | Cannot create intermediate directories. |
| Write failure | Cannot write the file to disk. |

# Private

## Transport constraints

All MCP tool parameters are transmitted as JSON. The
`content` parameter is a JSON string, which means it is
UTF-8 text. Binary files cannot be written through this
tool. This is consistent with the framework operating
exclusively on text files (see CODE_FROM_SPEC.md).
