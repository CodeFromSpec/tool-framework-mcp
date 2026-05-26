<!-- code-from-spec: ROOT/functional/mcp_tools/write_file@CfijaTdhezcP77uWhSKmHTpwIas -->

# WriteFile — Functional Pseudocode

## Data Structures

```
record Output
  id: string
  path: string

record Frontmatter
  outputs: list of Output
  (other fields omitted — only outputs is required here)
```

## Functions

---

### WriteFile(logical_name, path, content) -> string

Parameters:
- logical_name: string — the ROOT/ reference identifying the node whose outputs authorize this write
- path:         string — relative file path from project root
- content:      string — complete file content (UTF-8 text)

Returns: success message string

---

**Step 1 — Validate the logical name.**

  Call ValidateLogicalName(logical_name).
  If logical_name does not start with "ROOT/", or is otherwise not a recognized ROOT/ reference,
    raise error "invalid logical name: <logical_name> is not a recognized ROOT/ reference".

**Step 2 — Resolve the node path from the logical name.**

  Call ResolvePath(logical_name) to get the node's file path (e.g., "code-from-spec/x/y/_node.md").
  Strip any parenthetical qualifier from logical_name before resolving.

**Step 3 — Read and parse the node's frontmatter.**

  Call ParseFrontmatter(node_file_path).
  If the file cannot be read,
    raise error "invalid logical name: node file not found for <logical_name>".
  If the parsed frontmatter has an empty outputs list,
    raise error "no outputs: node <logical_name> has no outputs declared".

**Step 4 — Validate the write path for safety.**

  Call ValidatePath(path, project_root).
  If path is empty,
    raise error "path validation failure: path is empty".
  If path is absolute (starts with "/" or a drive letter like "C:"),
    raise error "path validation failure: path must be relative, not absolute".
  If path contains ".." components after normalization,
    raise error "path validation failure: directory traversal is not allowed".
  If the fully resolved absolute path falls outside the project root,
    raise error "path validation failure: path escapes the project root".

**Step 5 — Confirm path is declared in the node's outputs.**

  For each output entry in frontmatter.outputs:
    If output.path equals path,
      proceed to step 6.
  If no match was found after checking all entries,
    raise error "path not in outputs: <path> is not declared in the outputs of <logical_name>".

**Step 6 — Create intermediate directories if needed.**

  Determine the directory portion of path (all segments except the final filename).
  If the directory does not exist,
    attempt to create it, including any missing ancestor directories.
    If directory creation fails,
      raise error "directory creation failure: could not create directories for <path>".

**Step 7 — Write the file.**

  Write content to path (relative to project root), encoded as UTF-8.
  If the file already exists, overwrite it without warning.
  If writing fails for any reason (permissions, disk full, etc.),
    raise error "write failure: could not write to <path>".

**Step 8 — Return success.**

  Return "wrote <path>".

---

## Error Conditions Summary

| Condition                | When it occurs                                                            |
|--------------------------|---------------------------------------------------------------------------|
| invalid logical name     | logical_name is not a valid ROOT/ reference, or its node file is missing  |
| no outputs               | the node's frontmatter has no outputs field or the list is empty          |
| path validation failure  | path is empty, absolute, contains "..", or resolves outside project root  |
| path not in outputs      | path does not appear in the node's declared outputs list                  |
| directory creation failure | intermediate directories cannot be created                              |
| write failure            | the file cannot be written to disk                                        |

---

## Contracts

- Only writes to paths that are explicitly declared in the target node's outputs.
- Each call writes exactly one file.
- Intermediate directories are created automatically as needed.
- Existing files are overwritten without warning.
- The tool is stateless: every call independently resolves the node and validates inputs.
