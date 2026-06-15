<!-- code-from-spec: SPEC/functional/logic/mcp_tools/write_file@44gaq70qsiJr7pV3QE43WVgDNu8 -->

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
     Call FileWrite with that PathCfs and content.
     If it fails, propagate the error.

  8. Return "wrote <path>" where <path> is the path string.
