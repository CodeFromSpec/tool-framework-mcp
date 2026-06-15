<!-- code-from-spec: SPEC/functional/logic/os/file_writer@QUoqjVaMGCEnbSKgdZEUWwTI0Lc -->

function FileWrite(cfs_path: pathutils.PathCfs, content: string)

  1. Call PathCfsToOs with cfs_path to obtain an OS-native absolute path.
     If PathCfsToOs raises any error, propagate it to the caller unchanged.

  2. Derive the parent directory path from the resolved OS path.

  3. If the parent directory does not exist, create it along with any
     missing intermediate directories.
     If directory creation fails, raise error "CannotCreateDirectory".

  4. Write content to the file at the resolved OS path, encoded as UTF-8.
     If the file already exists, overwrite it completely.
     If writing fails for any reason, raise error "CannotWriteFile".
