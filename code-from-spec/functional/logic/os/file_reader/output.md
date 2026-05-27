<!-- code-from-spec: ROOT/functional/logic/os/file_reader@Sv32IjHNPt2iAnKhYCwdFeUMGRI -->

## Records

```
record FileReader
  cfs_path: CfsPath
  os_path:  OsPath
  stream:   file stream handle (optional)
  closed:   boolean
```


## Functions


### FileOpen(cfs_path) -> FileReader

  1. Call ResolvePath(cfs_path) to obtain an OsPath.
     If ResolvePath raises any error, propagate it to the caller unchanged.

  2. Attempt to open the file at the resolved OsPath for sequential reading.
     If the file cannot be opened (e.g. it does not exist or permissions deny
     access), raise error "file unreadable".

  3. Return a FileReader record with:
     - cfs_path set to the given cfs_path
     - os_path set to the resolved OsPath
     - stream set to the opened file stream
     - closed set to false


### FileReadLine(reader) -> string

  1. If reader.closed is true, raise error "end of file".

  2. Attempt to read the next line from reader.stream.
     If there are no more lines (stream is at end of file), raise error "end of file".

  3. Strip the trailing line terminator from the line.
     If the terminator is CRLF ("\r\n"), remove both characters.
     If the terminator is LF ("\n"), remove it.
     If there is no terminator (final line), leave the content as-is.

  4. Return the resulting string.


### FileSkipLines(reader, count)

  1. If reader.closed is true, return immediately (do nothing).

  2. Repeat count times:
       Attempt to read the next line from reader.stream.
       If the stream is at end of file, stop iterating and return.

  3. Return.


### FileClose(reader)

  1. If reader.closed is true, return immediately (do nothing).

  2. Release reader.stream (close the file handle).

  3. Set reader.closed to true.
     Set reader.stream to absent/empty.
