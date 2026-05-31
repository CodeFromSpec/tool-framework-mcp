<!-- code-from-spec: ROOT/functional/logic/os/file_writer@xqJ6IL-tpJa5ud7SoXAFazhoZpY -->

function FileWrite(cfs_path: pathutils.PathCfs, content: string)

  1. Convert cfs_path to an OS path by calling PathCfsToOs(cfs_path).
     If PathCfsToOs raises any error, propagate it to the caller unchanged.
     Let os_path be the resulting PathOs.

  2. Determine the parent directory of os_path.
     If the parent directory does not exist, create it and all
     intermediate directories required to make it exist.
     If any directory cannot be created, raise error "CannotCreateDirectory".

  3. Write content to the file at os_path, encoded as UTF-8 text.
     If the file already exists, overwrite it completely.
     If the file does not exist, create it.
     Write content exactly as received — no line-ending normalization
     or other transformation.
     If the write fails for any reason, raise error "CannotWriteFile".
