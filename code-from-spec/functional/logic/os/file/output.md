<!-- code-from-spec: SPEC/functional/logic/os/file@jWU0Bn-dNYVVrKxUntjows2G_3A -->

namespace: filereader

## Records

record FileHandle
  mode: string
  os_path: pathutils.PathOs
  stream: file stream handle
  closed: boolean
  line_buffer: buffered line reader

---

## Functions

function FileOpen(cfs_path: pathutils.PathCfs, mode: string) -> FileHandle

  1. If mode is not "read", "overwrite", or "append",
     raise error "InvalidMode".

  2. Call PathCfsToOs(cfs_path).
     If PathCfsToOs raises a PathUtils error, propagate it.

  3. If mode is "read":
       Open the file at the resulting PathOs for sequential reading,
       acquiring a shared lock.
       If the file cannot be opened (does not exist, permission
       denied, or other OS error), raise error "FileUnreadable".

  4. If mode is "overwrite":
       Create all intermediate directories leading to the file.
       If any directory cannot be created, raise error "CannotCreateDirectory".
       Open the file at the resulting PathOs for writing,
       truncating any existing content, acquiring an exclusive lock.
       If the file cannot be opened, raise error "CannotOpenFile".

  5. If mode is "append":
       Create all intermediate directories leading to the file.
       If any directory cannot be created, raise error "CannotCreateDirectory".
       Open the file at the resulting PathOs for writing
       without truncating (append to existing content),
       acquiring an exclusive lock.
       If the file cannot be opened, raise error "CannotOpenFile".

  6. Return a FileHandle record with:
     - mode set to mode
     - os_path set to the resolved PathOs
     - stream set to the opened file stream handle
     - closed set to false
     - line_buffer set to a buffered line reader wrapping the stream
       (only meaningful when mode is "read")


function FileReadLine(handle: FileHandle) -> string

  1. If handle.mode is not "read", raise error "WrongMode".

  2. If handle.closed is true, raise error "EndOfFile".

  3. Read the next line from handle.line_buffer up to and including
     the next newline character, or until end of stream.
     If there are no more bytes to read, raise error "EndOfFile".

  4. Strip the trailing line terminator from the line:
     If the line ends with "\r\n", remove both characters.
     Else if the line ends with "\n", remove that character.

  5. Return the resulting string.


function FileWrite(handle: FileHandle, content: string)

  1. If handle.mode is not "overwrite" and handle.mode is not "append",
     raise error "WrongMode".

  2. Write content to handle.stream encoded as UTF-8, exactly as
     received with no transformation of line endings or other content.
     If the write fails for any reason, raise error "CannotWriteFile".


function FileSkipLines(handle: FileHandle, count: integer)

  1. If handle.mode is not "read", raise error "WrongMode".

  2. If handle.closed is true, return immediately.

  3. Repeat count times:
       Read and discard the next line from handle.line_buffer
       (up to and including the next newline, or end of stream).
       If end of stream is reached before completing all
       iterations, stop iterating without raising an error.


function FileClose(handle: FileHandle)

  1. If handle.closed is true, return immediately.

  2. Release handle.stream (close the OS file handle and release the lock).

  3. Set handle.closed to true.


function FileRename(source: pathutils.PathCfs, destination: pathutils.PathCfs)

  1. Call PathCfsToOs(source).
     If PathCfsToOs raises a PathUtils error, propagate it.
     Store the result as source_os.

  2. Call PathCfsToOs(destination).
     If PathCfsToOs raises a PathUtils error, propagate it.
     Store the result as destination_os.

  3. Perform an atomic OS-level rename of source_os to destination_os.
     If the destination already exists, overwrite it.
     If the rename fails for any reason, raise error "CannotRename".


function FileDelete(cfs_path: pathutils.PathCfs)

  1. Call PathCfsToOs(cfs_path).
     If PathCfsToOs raises a PathUtils error, propagate it.

  2. Delete the file at the resulting PathOs.
     If the file cannot be deleted (does not exist, permission
     denied, or other OS error), raise error "CannotDelete".
