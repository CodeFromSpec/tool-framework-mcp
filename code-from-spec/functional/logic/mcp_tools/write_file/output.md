<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@aVMNbZhlzwCoFf-zJI3wCRlmgvQ -->

function MCPWriteFile(logical_name: string, path: string, content: string) -> string
  errors:
    - QualifierNotAllowed: the logical name contains a parenthetical qualifier.
    - UnreadableFrontmatter: the node's frontmatter cannot be parsed.
    - NoOutput: target node has no output field.
    - PathNotInOutput: path is not declared in the node's output.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileWriter.*): propagated from FileWrite.

  1. Call LogicalNameHasQualifier(logical_name).
     If true, raise error "qualifier not allowed".

  2. Call LogicalNameToPath(logical_name) -> node_path.
     If it fails, propagate the error.

  3. Call FrontmatterParse(node_path) -> frontmatter.
     If it fails, raise error "unreadable frontmatter".

  4. If frontmatter.output is empty, raise error "no output".

  5. Call PathValidateCfs(path).
     If it fails, propagate the error.

  6. Compare path against frontmatter.output using exact string match.
     If they do not match, raise error "path not in output".

  7. Construct a PathCfs record with value set to path.
     Call FileWrite(PathCfs, content).
     If it fails, propagate the error.

  8. Return "wrote <path>" where <path> is the path string.
