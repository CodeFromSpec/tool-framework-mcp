<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@RrKapyhhJqUl_20_vvhym22eYDs -->

## Namespace

    namespace: frontmatter

## Records

```
record FrontmatterExternal
  path: string

record FrontmatterOutput
  id: string
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of FrontmatterExternal
  input: string
  outputs: list of FrontmatterOutput
```

## Functions

---

function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter

  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters is not valid YAML,
      or a required field within a sub-record is missing.
    - (FileReader.*): propagated from FileOpen.

  1. Call FileOpen(file_path) to obtain a reader.
     If FileOpen raises FileUnreadable or any PathUtils error, propagate it.

  2. Call FileReadLine(reader) to read the first line.
     If it raises EndOfFile, call FileClose(reader) and return an empty
     Frontmatter record.
     If the first line is not exactly "---", call FileClose(reader) and
     return an empty Frontmatter record.

  3. Collect subsequent lines into a buffer until a line that is exactly
     "---" is found or EndOfFile is raised.

     For each line:
       a. Call FileReadLine(reader).
       b. If it raises EndOfFile, call FileClose(reader) and raise
          MalformedYAML "missing closing --- delimiter".
       c. If the line is exactly "---", stop collecting and proceed to
          step 4.
       d. Otherwise, append the line to the buffer.

  4. Call FileClose(reader).

  5. If the buffer is empty, return an empty Frontmatter record.

  6. Parse the buffer as YAML.
     If YAML parsing fails, raise MalformedYAML "invalid YAML in
     frontmatter block".

  7. Extract fields from the parsed YAML, ignoring unknown keys:

     depends_on:
       If present, read as a list of strings.
       If absent, use an empty list.

     external:
       If present, read as a list of mappings.
       For each entry:
         - Read "path" as a string.
         - If "path" is absent or empty, raise MalformedYAML
           "external entry missing required field: path".
         - Construct a FrontmatterExternal record with that path.
       If absent, use an empty list.

     input:
       If present, read as a string.
       If absent, use an empty string.

     outputs:
       If present, read as a list of mappings.
       For each entry:
         - Read "id" as a string.
         - If "id" is absent or empty, raise MalformedYAML
           "outputs entry missing required field: id".
         - Read "path" as a string.
         - If "path" is absent or empty, raise MalformedYAML
           "outputs entry missing required field: path".
         - Construct a FrontmatterOutput record with id and path.
       If absent, use an empty list.

  8. Return a Frontmatter record with the extracted fields.
```
