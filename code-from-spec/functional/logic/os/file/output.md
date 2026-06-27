<!-- code-from-spec: SPEC/functional/logic/os/file_reader@ZwniNt1NJsZMAGofjefSjhfWfyM -->

namespace: filereader

## Records

record FileReader
  cfs_path: pathutils.PathCfs
  os_path: pathutils.PathOs
  stream: file stream handle
  closed: boolean

---

## Functions

function FileOpen(cfs_path: pathutils.PathCfs) -> FileReader

  1. Call PathCfsToOs(cfs_path).
     If PathCfsToOs raises a PathUtils error, propagate it.

  2. Open the file at the resulting PathOs for sequential reading
     from the beginning.
     If the file cannot be opened (does not exist, permission
     denied, or other OS error), raise error "FileUnreadable".

  3. Return a FileReader record with:
     - cfs_path set to cfs_path
     - os_path set to the resolved PathOs
     - stream set to the opened file stream handle
     - closed set to false


function FileReadLine(reader: FileReader) -> string

  1. If reader.closed is true, raise error "EndOfFile".

  2. Read the next line from reader.stream up to and including
     the next newline character, or until end of stream.
     If there are no more bytes to read, raise error "EndOfFile".

  3. Strip the trailing line terminator from the line:
     If the line ends with "\r\n", remove both characters.
     Else if the line ends with "\n", remove that character.

  4. Return the resulting string.


function FileSkipLines(reader: FileReader, count: integer)

  1. If reader.closed is true, return immediately.

  2. Repeat count times:
       Read and discard the next line from reader.stream
       (up to and including the next newline, or end of stream).
       If end of stream is reached before completing all
       iterations, stop iterating without raising an error.


function FileClose(reader: FileReader)

  1. If reader.closed is true, return immediately.

  2. Release reader.stream (close the OS file handle).

  3. Set reader.closed to true.
