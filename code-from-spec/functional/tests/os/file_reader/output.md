<!-- code-from-spec: ROOT/functional/tests/os/file_reader@cnJu1DD18EWvwvKwNd_YLzklnro -->

## Test cases

### Happy path

#### Opens and reads all lines

Setup: Create a file containing three lines: `"alpha"`, `"beta"`, `"gamma"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `"beta"`.
4. Call `FileReadLine` — expect `"gamma"`.
5. Call `FileReadLine` — expect EndOfFile.

#### Normalizes CRLF to LF

Setup: Create a file containing `"alpha"` and `"beta"` with CRLF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"` (no CR or LF characters).
3. Call `FileReadLine` — expect `"beta"` (no CR or LF characters).

#### Reads file with no trailing newline

Setup: Create a file containing `"alpha"` (with LF) and `"beta"` (no trailing newline).

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `"beta"`.
4. Call `FileReadLine` — expect EndOfFile.

#### FileSkipLines advances the reader

Setup: Create a file containing `"one"`, `"two"`, `"three"`, `"four"`, `"five"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count 2.
3. Call `FileReadLine` — expect `"three"`.

#### FileSkipLines past end of file

Setup: Create a file containing `"one"`, `"two"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileSkipLines` with count 10 — expect no error.
3. Call `FileReadLine` — expect EndOfFile.

#### Preserves leading whitespace

Setup: Create a file containing `"  alpha"` and `"    beta"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"  alpha"`.
3. Call `FileReadLine` — expect `"    beta"`.

#### Preserves trailing whitespace

Setup: Create a file containing `"alpha  "` and `"beta   "` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha  "`.
3. Call `FileReadLine` — expect `"beta   "`.

#### Preserves internal whitespace

Setup: Create a file containing `"alpha   beta"` and `"one\ttwo"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha   beta"`.
3. Call `FileReadLine` — expect `"one\ttwo"`.

#### Preserves empty lines

Setup: Create a file containing `"alpha"`, `""`, `""`, `"beta"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"alpha"`.
3. Call `FileReadLine` — expect `""`.
4. Call `FileReadLine` — expect `""`.
5. Call `FileReadLine` — expect `"beta"`.

#### Preserves non-ASCII characters

Setup: Create a file containing `"café"`, `"日本語"`, `"🎉🚀"` with LF endings.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"café"`.
3. Call `FileReadLine` — expect `"日本語"`.
4. Call `FileReadLine` — expect `"🎉🚀"`.

### Edge cases

#### Empty file

Setup: Create an empty file.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect EndOfFile immediately.

#### Single line without newline

Setup: Create a file containing only `"hello"` with no newline.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileReadLine` — expect `"hello"`.
3. Call `FileReadLine` — expect EndOfFile.

### Failure cases

#### File does not exist

Setup: No file is created; use a path that does not exist.

Actions:
1. Call `FileOpen` with the non-existent path — expect FileUnreadable.

#### Read after close

Setup: Create a file containing `"alpha"` with LF ending.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileReadLine` — expect EndOfFile.

#### Skip after close

Setup: Create a file containing `"alpha"` with LF ending.

Actions:
1. Call `FileOpen` with the file path.
2. Call `FileClose`.
3. Call `FileSkipLines` with count 1 — expect no error, the call does nothing.
