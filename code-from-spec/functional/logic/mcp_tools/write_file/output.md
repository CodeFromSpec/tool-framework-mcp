<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@noYOQhUqUQBWZT48nEehc4IogS0 -->

function MCPWriteFile(logical_name: string, path: string, content: string) -> string

  1. Call LogicalNameToPath(logical_name) to get the node's file path.
     If it fails, propagate the error.

  2. Call FrontmatterParse(node_path) to read the node's frontmatter.
     If parsing fails, raise error "unreadable frontmatter".

  3. If frontmatter.outputs is empty, raise error "no outputs".

  4. Call PathValidateCfs(path) to validate the path string.
     If it fails, propagate the error.

  5. For each entry in frontmatter.outputs:
       If entry.path equals path (exact string match):
         proceed to step 6.
     If no match was found, raise error "path not in outputs".

  6. Construct a PathCfs record with value set to path.
     Call FileWrite(cfs_path, content).
     If it fails, propagate the error.

  7. Return "wrote <path>" where <path> is the path string.
