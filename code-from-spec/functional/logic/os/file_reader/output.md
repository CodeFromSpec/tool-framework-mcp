<!-- code-from-spec: ROOT/functional/logic/os/file_reader@JnDhXaXPtHJJuac5F21bT0mXuBo -->

function FileOpen(cfs_path: pathutils.PathCfs) -> FileReader
  errors:
    - FileUnreadable
    - (PathUtils.*): propagated from PathCfsToOs

  1. Call PathCfsToOs(cfs_path).
     If it raises a PathUtils error, propagate it.
     Let os_path be the result.

  2. Open the file at os_path for sequential reading from the beginning.
     If the file cannot be opened (does not exist, permission denied,
     or any other OS error), raise error "FileUnreadable".

  3. Return a FileReader record with:
     - cfs_path: cfs_path
     - an internal stream handle pointing to the beginning of the file
     - a flag indicating the stream is open


function FileReadLine(reader: FileReader) -> string
  errors:
    - EndOfFile: no more lines to read

  1. If the reader's stream is closed, raise error "EndOfFile".

  2. Read bytes from the stream up to and including the next LF character,
     or until the stream is exhausted, whichever comes first.
     Do not buffer more than the current line in memory.

  3. If no bytes were read (stream was already at end), raise error "EndOfFile".

  4. Let line be the bytes read.
     If line ends with CRLF, strip the trailing CR and LF.
     Else if line ends with LF, strip the trailing LF.

  5. Return line as a string.


function FileSkipLines(reader: FileReader, count: integer)

  1. If the reader's stream is closed, return immediately.

  2. Repeat count times:
       Call FileReadLine(reader).
       If it raises "EndOfFile", stop repeating and return.


function FileClose(reader: FileReader)

  1. If the reader's stream is already closed, return immediately.

  2. Release the OS file handle associated with the reader's stream.

  3. Mark the reader's stream as closed.
