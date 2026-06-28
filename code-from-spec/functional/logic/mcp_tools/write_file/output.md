<!-- code-from-spec: SPEC/functional/logic/mcp_tools/write_file@y-fi8kgCcVX9N_ar3rYxLcQoPVo -->

function MCPWriteFile(logical_name: string, path: string, content: string) -> string

  1. Call LogicalNameHasQualifier with logical_name.
     If true, raise error "qualifier not allowed".

  2. Call LogicalNameToPath with logical_name.
     If it fails, propagate the error.
     Store the result as node_path.

  3. Call FrontmatterParse with node_path.
     If it fails, raise error "unreadable frontmatter".
     Store the result as frontmatter.

  4. If frontmatter.output is empty, raise error "no output".

  5. Call PathValidateCfs with path.
     If it fails, propagate the error.

  6. If path does not exactly match frontmatter.output,
     raise error "path not in output".

  7. Construct a PathCfs record with value set to path.
     Call FileOpen with that PathCfs, mode "overwrite", and timeout 30000.
     If it fails, propagate the error.
     Store the result as handle.

  8. Call FileWrite with handle and content.
     If it fails, call FileClose with handle, then propagate the error.

  9. Call FileClose with handle.

  10. Return "wrote <path>" where <path> is the path string.
