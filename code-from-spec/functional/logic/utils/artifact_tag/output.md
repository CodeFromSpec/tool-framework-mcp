<!-- code-from-spec: ROOT/functional/utils/artifact_tag@lkuRRj0KR4okwRUT9yyhea0u9eY -->

## interface

```
record FileReader
  file_path: string

function OpenFileReader(file_path) -> FileReader
  errors:
    - file unreadable: the file cannot be opened.

function ReadLine(reader) -> line
  errors:
    - end of file: no more lines to read.

function SkipLines(reader, count)

function Close(reader)
```

`OpenFileReader` opens a file and prepares it for
sequential line-by-line reading. The file remains open
until `Close` is called.

`ReadLine` reads the next line from the file, normalizes
CRLF to LF, and returns the line without the terminator.
Raises "end of file" when there are no more lines.

`SkipLines` reads and discards `count` lines without
returning their content.

`Close` releases the file resource. After `Close`, any
call to `ReadLine` or `SkipLines` raises "end of file".
