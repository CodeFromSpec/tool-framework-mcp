<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@1Rb_j6ZEv_liolvA_-T2bkY7dPw -->

# MCPWriteFile

## Function signature

```
function MCPWriteFile(logical_name: string, path: string, content: string) -> string
  errors:
    - UnreadableFrontmatter: the node's frontmatter cannot be parsed.
    - NoOutputs: target node has no outputs field.
    - PathNotInOutputs: path is not declared in the node's outputs.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileWriter.*): propagated from FileWrite.
```

## Logic

### Step 1 — Read frontmatter

1. Call `LogicalNameToPath` with `logical_name`.
   If it raises an error, propagate it unchanged.

2. Call `FrontmatterParse` with the resolved path.
   If it raises an error, raise error "unreadable frontmatter".

3. If `frontmatter.outputs` is empty, raise error "no outputs".

### Step 2 — Validate path

4. Construct a `PathCfs` record with `value` set to `path`.

5. Call `PathValidateCfs` with the `path` string.
   If it raises an error, propagate it unchanged.

### Step 3 — Check path against outputs

6. For each entry in `frontmatter.outputs`:
     If `entry.path` equals `path` (exact string match), proceed to Step 4.

7. If no entry matched, raise error "path not in outputs".

### Step 4 — Write file

8. Call `FileWrite` with the `PathCfs` and `content`.
   If it raises an error, propagate it unchanged.

9. Return `"wrote <path>"` where `<path>` is the value of the `path` parameter.

## Error conditions

| Error | Condition |
|---|---|
| `UnreadableFrontmatter` | `FrontmatterParse` fails for any reason. |
| `NoOutputs` | The node's frontmatter has an empty `outputs` list. |
| `PathNotInOutputs` | The `path` string does not exactly match the `path` field of any entry in `frontmatter.outputs`. |
| `LogicalNames.*` | Propagated from `LogicalNameToPath` (e.g. `UnsupportedReference`). |
| `PathUtils.*` | Propagated from `PathValidateCfs` (e.g. `PathEmpty`, `PathAbsolute`, `PathContainsBackslash`, `DirectoryTraversal`). |
| `FileWriter.*` | Propagated from `FileWrite` (e.g. `CannotCreateDirectory`, `CannotWriteFile`). |
