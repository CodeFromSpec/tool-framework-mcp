<!-- code-from-spec: ROOT/functional/logic/os/file_writer@q-3YkwzF852pRRy5itHEhAj2UrA -->

# FileWrite

function FileWrite(cfs_path: PathCfs, content: string)

  1. Convert cfs_path to an OS path by calling PathCfsToOs(cfs_path).
     If PathCfsToOs raises an error, propagate it unchanged.
     Assign the result to os_path.

  2. Determine the parent directory of os_path.
     If the parent directory does not exist, create it along
     with all intermediate directories.
     If any directory cannot be created, raise error
     "cannot create directory".

  3. Write content to the file at os_path, encoded as UTF-8.
     If the file already exists, overwrite it without warning.
     If the file cannot be written, raise error "cannot write file".

  4. Return with no value (write succeeded).
