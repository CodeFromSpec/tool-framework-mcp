<!-- code-from-spec: ROOT/functional/logic/os/file_writer@IewsWrHBC7h76zYZS2jFrqp87pI -->

function FileWrite(cfs_path: PathCfs, content: string)

  1. Call PathCfsToOs(cfs_path) to obtain an OS-native absolute path.
     If PathCfsToOs raises any error, propagate it to the caller.

  2. Determine the parent directory of the resolved OS path.

  3. If the parent directory does not exist, create it and all
     intermediate directories recursively.
     If creation fails, raise error "CannotCreateDirectory".

  4. Write content to the file at the resolved OS path, encoded as UTF-8.
     If the file already exists, overwrite it without warning.
     If writing fails, raise error "CannotWriteFile".
