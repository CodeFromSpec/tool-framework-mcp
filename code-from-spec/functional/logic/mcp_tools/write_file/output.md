<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@EHwM0Hvwvn-wVVr0Zr7k2u-j-6Q -->

# MCPWriteFile

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

## Step 1 — Read frontmatter

1. Call `LogicalNameToPath` with `logical_name`.
   If it raises an error, propagate it to the caller.

2. Call `FrontmatterParse` with the resolved path.
   If parsing fails, raise error "unreadable frontmatter".

3. If `frontmatter.outputs` is empty, raise error "no outputs".

## Step 2 — Validate path

4. Construct a `PathCfs` record with `value` set to `path`.

5. Call `PathValidateCfs` with `path`.
   If it raises an error, propagate it to the caller.

## Step 3 — Check path against outputs

6. For each entry in `frontmatter.outputs`:
     If `entry.path` equals `path` (exact string match), proceed to Step 4.

7. If no entry matched, raise error "path not in outputs".

## Step 4 — Write file

8. Call `FileWrite` with the `PathCfs` record and `content`.
   If it raises an error, propagate it to the caller.

9. Return `"wrote <path>"` where `<path>` is the `path` string.

## Contracts

- Only writes to paths declared in the node's `outputs` list.
- Path validation uses `PathValidateCfs` — same rules as the OS layer.
- File writing uses `FileWrite` — creates intermediate directories,
  overwrites existing files.
- Writes exactly one file per call.
- Content is written exactly as received — no line ending normalization
  or other transformations.
