---
depends_on:
  - ARTIFACT/golang/interfaces/mcp_tools/write_file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/utils/logical_names
output: internal/mcpwritefile/mcpwritefile.go
---

# SPEC/golang/implementation/mcp_tools/write_file

# Agent

Implement the write file tool as a Go package.

## Logic

1. Call `LogicalNameHasQualifier` with logical_name.
   If true, return error "qualifier not allowed".

2. Call `LogicalNameToPath` with logical_name. If it
   fails, propagate the error. Store the result as
   node_path.

3. Call `FrontmatterParse` with node_path. If it fails,
   return error "unreadable frontmatter". Store the
   result as frontmatter.

4. If `frontmatter.output` is empty, return error
   "no output".

5. Call `PathValidateCfs` with path. If it fails,
   propagate the error.

6. If path does not exactly match `frontmatter.output`,
   return error "path not in output".

7. Construct a `PathCfs` record with value set to path.
   Call `FileOpen` with that PathCfs, mode "overwrite",
   and timeout 30000. If it fails, propagate the error.
   Store the result as handle.

8. Call `FileWrite` with handle and content. If it
   fails, call `FileClose` with handle, then propagate
   the error.

9. Call `FileClose` with handle.

10. Return "wrote <path>" where <path> is the path
    string.

## Go-specific guidance

- Use the `logicalnames` package for `LogicalNameToPath`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `file` package for `FileOpen`, `FileWrite`,
  `FileClose`.
- The package name should be `mcpwritefile`.
- The function receives plain strings from the MCP
  transport layer. Construct `PathCfs` internally.
