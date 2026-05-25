<!-- code-from-spec: ROOT/functional/mcp_tools/write_file@PENDING -->

## Functions

### function WriteFile(logical_name, path, content) -> string

Writes a generated source file to disk after validating the path
against the node's declared outputs.

**Parameters**

- logical_name: string, required. Logical name of the node whose
  outputs authorize the write.
- path: string, required. Relative file path from project root.
- content: string, required. Complete file content as UTF-8 text.

**Validation**

1. Validate logical_name as a ROOT/ reference using logical_names.
   If invalid, raise error "invalid logical name".

2. Resolve logical_name to a file path using logical_names.
   Read the frontmatter of that file using frontmatter parser.
   If the file cannot be read, raise error "unreadable file".

3. If frontmatter has no outputs field, raise error "no outputs".

4. Validate the path:
   a. If path is empty, raise error "path validation failure".
   b. If path is absolute (starts with "/" or a drive letter),
      raise error "path validation failure".
   c. Normalize the path and check for ".." components.
      If any are present, raise error "path validation failure".
   d. Resolve the full path by joining the project root with path.
      Resolve any symlinks. If the resolved path is outside the
      project root, raise error "path validation failure".

5. Check that path appears in the node's outputs list by comparing
   it against the path field of each output entry.
   If path is not found, raise error "path not in outputs".

**Write**

6. Determine the directory portion of the full file path.
   If the directory does not exist, create it and all intermediate
   directories.
   If directory creation fails, raise error "directory creation failure".

7. Write content to the file at the full path.
   If the file already exists, overwrite it.
   If writing fails, raise error "write failure".

8. Return "wrote <path>" where <path> is the relative path provided.
