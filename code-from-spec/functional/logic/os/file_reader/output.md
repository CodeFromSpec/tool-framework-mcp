<!-- code-from-spec: ROOT/functional/logic/os/file_reader@ULJTiTTBoii6ySFSJTihE0RAouQ -->

namespace: filereader

---

record FileReader
  cfs_path: pathutils.PathCfs
  os_path:  pathutils.PathOs
  stream:   open file stream (sequential, forward-only)
  closed:   boolean

---

function FileOpen(cfs_path: pathutils.PathCfs) -> FileReader
  errors:
    - FileUnreadable
    - (PathUtils.*): propagated from PathCfsToOs

  1. Call PathCfsToOs(cfs_path).
     If PathCfsToOs raises an error, propagate it unchanged.
     Let os_path be the returned PathOs.

  2. Open a sequential read stream at os_path.
     If the stream cannot be opened for any reason
     (file not found, permission denied, or other OS error),
     raise error "FileUnreadable".

  3. Return a FileReader record with:
       cfs_path = cfs_path
       os_path  = os_path
       stream   = the opened stream
       closed   = false

---

function FileReadLine(reader: FileReader) -> string
  errors:
    - EndOfFile: no more lines to read

  1. If reader.closed is true, raise error "EndOfFile".

  2. Read the next sequence of bytes from reader.stream
     up to and including the next LF character, or until
     the stream is exhausted, whichever comes first.
     Do not load the entire file into memory — consume
     only the bytes needed for this one line.

  3. If no bytes were read (the stream was already at the end),
     raise error "EndOfFile".

  4. Let line be the collected bytes decoded as text.

  5. If line ends with CRLF, strip the trailing CR and LF.
     If line ends with LF only, strip the trailing LF.
     If line has no trailing newline (final line of file),
     return it as-is.

  6. Return line.

---

function FileSkipLines(reader: FileReader, count: integer)

  1. If reader.closed is true, return immediately.

  2. Repeat count times:
       Call FileReadLine(reader).
       If FileReadLine raises "EndOfFile", stop iterating
       and return without error.

  3. Return.

---

function FileClose(reader: FileReader)

  1. If reader.closed is true, return immediately.

  2. Release reader.stream (close the OS file handle).

  3. Set reader.closed = true.

  4. Return.
