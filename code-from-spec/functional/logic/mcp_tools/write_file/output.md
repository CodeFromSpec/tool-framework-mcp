<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@87JEQZ7dhsfoSGYzMqDCdqtkH44 -->

## Functions

function MCPWriteFile(logical_name: string, path: string, content: string) -> string
  errors:
    - UnreadableFrontmatter: the node's frontmatter cannot be parsed.
    - NoOutput: target node has no output field.
    - PathNotInOutput: path is not declared in the node's output.
    - (LogicalNames.*): propagated from LogicalNameToPath.
    - (PathUtils.*): propagated from PathValidateCfs.
    - (FileWriter.*): propagated from FileWrite.

  1. Read frontmatter.

     Call LogicalNameToPath with logical_name. If it fails, propagate the error.

     Call FrontmatterParse with the resolved file path.
     If parsing fails, raise error "unreadable frontmatter".

     If frontmatter.output is empty, raise error "no output".

  2. Validate path.

     Call PathValidateCfs on the path string. If it fails, propagate the error.

  3. Check path against output.

     Compare path against frontmatter.output using exact string match.
     If they do not match, raise error "path not in output".

  4. Write file.

     Construct a PathCfs from the path string.
     Call FileWrite with the PathCfs and content. If it fails, propagate the error.

     Return "wrote <path>" where <path> is the path string.
