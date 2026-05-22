---
outputs:
  - id: write_file
    path: code-from-spec/functional/tools/write_file/output.md
---

# ROOT/functional/tools/write_file

Writes a generated source file to disk after validating the
path against the node's declared outputs.

# Public

## Behavior

### Input

| Parameter | Required | Description |
|---|---|---|
| `logical_name` | yes | Logical name of the node whose outputs authorize the write. |
| `path` | yes | Relative file path from project root. |
| `content` | yes | Complete file content to write. |

### Output

A success message: `"wrote <path>"`.

### Validation

Before writing:
1. The logical name must be a valid `ROOT/` reference.
2. The target node must have `outputs` declared.
3. The path must pass path validation against the project root.
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
