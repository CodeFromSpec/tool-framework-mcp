<!-- code-from-spec: ROOT/functional/utils/file_reader@PENDING -->

record FileReader
  file_path: string
  lines: list of strings
  position: integer

function OpenFileReader(file_path) -> FileReader

  1. Open the file at file_path for reading.
     If the file cannot be opened, raise error "file unreadable".

  2. Read the entire contents of the file as text.

  3. Replace every occurrence of CRLF with LF.

  4. Split the resulting text on LF into a list of strings.
     If the text ends with LF, the split produces a trailing
     empty string — remove it. This ensures a final newline
     does not create a phantom empty line, while a file with
     no trailing newline still returns its last line.

  5. Create a FileReader record:
     - file_path: the original file_path
     - lines: the list from step 4
     - position: 0

  6. Return the FileReader.

function ReadLine(reader) -> line

  1. If reader.position is greater than or equal to the length
     of reader.lines, raise error "end of file".

  2. Let line be reader.lines at index reader.position.

  3. Advance reader.position by 1.

  4. Return line.

function SkipLines(reader, count)

  1. Advance reader.position by count.

  2. If reader.position exceeds the length of reader.lines,
     set reader.position to the length of reader.lines.
     This is not an error.
