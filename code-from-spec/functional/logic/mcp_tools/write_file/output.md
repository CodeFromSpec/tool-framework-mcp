<!-- code-from-spec: ROOT/functional/logic/mcp_tools/write_file@t6tcOhdJmq4g8acjsKlO3blmuPU -->

# MCPWriteFile

function MCPWriteFile(logical_name: string, path: string, content: string) -> string

  Parameters:
    - logical_name: logical name of the node whose outputs authorize the write
    - path: relative file path from project root (forward slashes)
    - content: complete file content (UTF-8 text)

  Returns: a success message string

  Errors:
    - (logical name errors): propagated from LogicalNameToPath
    - "unreadable frontmatter": the node's frontmatter cannot be parsed
    - "no outputs": target node has no outputs field or it is empty
    - "path not in outputs": path is not declared in the node's outputs
    - (path errors): propagated from PathValidateCfs
    - (write errors): propagated from FileWrite

  Steps:

  1. Resolve the logical name to a file path.
     Call LogicalNameToPath with logical_name.
     If it fails, propagate the error.

  2. Parse the node's frontmatter.
     Call FrontmatterParse with the resolved path from step 1.
     If parsing fails, raise error "unreadable frontmatter".

  3. Check that the node declares outputs.
     If frontmatter.outputs is empty, raise error "no outputs".

  4. Validate the path format.
     Call PathValidateCfs with the path string.
     If validation fails, propagate the error.

  5. Check that path is declared in the node's outputs.
     For each entry in frontmatter.outputs:
       If entry.path equals path (exact string match), proceed to step 6.
     If no matching entry is found, raise error "path not in outputs".

  6. Write the file.
     Construct a PathCfs from the path string.
     Call FileWrite with the PathCfs and content.
     If writing fails, propagate the error.

  7. Return "wrote <path>" where <path> is the path string.
