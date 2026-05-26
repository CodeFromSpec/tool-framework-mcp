# ROOT/golang/internal/file_reader

Sequential line reader for text files. Normalizes line
endings on read.

# Public

## Package

`package filereader`

## Interface

```go
type FileReader struct {
    // unexported fields
}

var (
    ErrOpen      = errors.New("file unreadable")
    ErrEndOfFile = errors.New("end of file")
)

func OpenFileReader(filePath string) (*FileReader, error)
func (r *FileReader) ReadLine() (string, error)
func (r *FileReader) SkipLines(count int)
func (r *FileReader) Close()
```

`OpenFileReader` opens a file and prepares it for sequential
line-by-line reading. Returns `ErrOpen` if the file cannot
be opened.

`ReadLine` returns the next line without the line terminator.
CRLF is normalized to LF before splitting. Returns
`ErrEndOfFile` when there are no more lines.

`SkipLines` advances the reader by `count` lines without
returning their content. Skipping past end-of-file is not
an error.

`Close` releases the underlying file handle. Callers must
call `Close` when done reading. After `Close`, `ReadLine`
returns `ErrEndOfFile`.

### Error handling

All errors wrap a sentinel so callers can use `errors.Is()`:

| Sentinel | Returned when |
|---|---|
| `ErrOpen` | The file cannot be opened. |
| `ErrEndOfFile` | No more lines to read. |
