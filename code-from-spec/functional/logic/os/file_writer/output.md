<!-- code-from-spec: ROOT/functional/logic/os/file_writer@fwciNcR78Bu3hFD1k-Mvx5DBnqo -->

function FileWrite(cfs_path: pathutils.PathCfs, content: string)

  1. Call PathCfsToOs(cfs_path) to get os_path.
     If PathCfsToOs raises any error, propagate it.

  2. Extract the parent directory from os_path.

  3. If the parent directory does not exist, create it and
     all intermediate directories.
     If creation fails, raise error "CannotCreateDirectory".

  4. Write content to the file at os_path, encoded as UTF-8.
     If the file already exists, overwrite it.
     If writing fails, raise error "CannotWriteFile".
