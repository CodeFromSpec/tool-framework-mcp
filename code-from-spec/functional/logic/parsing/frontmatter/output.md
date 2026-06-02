<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@OXHUGDYVXIfc5IcmSjvVUzcn49A -->

## Namespace

    namespace: frontmatter

## Records

```
record FrontmatterExternal
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of FrontmatterExternal
  input: string
  output: string
```

All fields default to empty (empty list, empty string) when absent from the YAML.

## Functions

```
function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters is not valid YAML,
      or a required sub-record field is missing.
    - (FileReader.*): propagated from FileOpen.
```

### FrontmatterParse

  1. Call `FileOpen` with `file_path`.
     If `FileOpen` raises `FileUnreadable` or a `PathUtils.*` error, propagate it.

  2. Read the first line using `FileReadLine`.
     If `FileReadLine` raises `EndOfFile`, call `FileClose` and return an empty
     `Frontmatter` record.
     If the first line is not exactly `"---"`, call `FileClose` and return an
     empty `Frontmatter` record.

  3. Collect lines into a buffer until a line that is exactly `"---"` is found.
     For each line read:
       If `FileReadLine` raises `EndOfFile` before the closing `"---"` is found,
       call `FileClose` and raise error `MalformedYAML`.
       If the line is exactly `"---"`, stop collecting.
       Otherwise, append the line to the buffer.

  4. Call `FileClose`.

  5. Parse the collected buffer as YAML.
     If parsing fails, raise error `MalformedYAML`.

  6. Extract known fields from the parsed YAML:
     - `depends_on`: list of strings. Default to empty list if absent.
     - `external`: list of records, each with a required `path` field (string).
       If any entry is missing the `path` field, raise error `MalformedYAML`.
       Default to empty list if absent.
     - `input`: string. Default to empty string if absent.
     - `output`: string. Default to empty string if absent.
     Silently ignore any other keys present in the YAML.

  7. Return a `Frontmatter` record with the extracted fields.
