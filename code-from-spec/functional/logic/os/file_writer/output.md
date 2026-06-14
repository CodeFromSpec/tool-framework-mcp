<!-- code-from-spec: ROOT/functional/logic/os/file_writer@BuPHKUHvfDYADvEv52YLURoS8uM -->

function FileWrite(cfs_path: pathutils.PathCfs, content: string)

  1. Call PathCfsToOs(cfs_path) to get the absolute OS path.
     If PathCfsToOs raises any error, propagate it to the caller.

  2. Derive the parent directory from the resolved OS path.

  3. If the parent directory does not exist, create it along with
     any missing intermediate directories.
     If creation fails, raise error "CannotCreateDirectory".

  4. Write content to the file at the resolved OS path, encoded
     as UTF-8 text.
     If the file already exists, overwrite it without warning.
     If writing fails, raise error "CannotWriteFile".
