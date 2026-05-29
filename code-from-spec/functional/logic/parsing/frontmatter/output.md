<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@1155pQz-TrURi9aRo5XFUb3-Wec -->

# FrontmatterParse

## Records

```
record FrontmatterExternalFragment
  description: optional string
  lines: string
  hash: string

record FrontmatterExternal
  path: string
  fragments: optional list of FrontmatterExternalFragment

record FrontmatterOutput
  id: string
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of FrontmatterExternal
  input: string
  outputs: list of FrontmatterOutput
```

## function FrontmatterParse(file_path: PathCfs) -> Frontmatter

  errors:
    - path errors: propagated from FileOpen
    - "file unreadable": the file cannot be opened or read
    - "malformed YAML": the content between --- delimiters is not valid YAML,
      or required sub-record fields are missing

  1. Call FileOpen(file_path) to obtain a FileReader.
     If FileOpen raises an error, propagate it to the caller.

  2. Read the first line using FileReadLine.
     If reading raises "end of file", call FileClose and return an empty Frontmatter record.
     If the first line is not exactly "---", call FileClose and return an empty Frontmatter record.

  3. Collect YAML lines:
     Create an empty list called yaml_lines.
     Loop:
       Read the next line using FileReadLine.
       If reading raises "end of file":
         Call FileClose.
         Raise error "malformed YAML".
       If the line is exactly "---":
         Stop collecting. The closing delimiter has been found.
       Otherwise:
         Append the line to yaml_lines.

  4. Call FileClose to release the file resource.

  5. If yaml_lines is empty, return an empty Frontmatter record.

  6. Join yaml_lines with newline characters to form yaml_text.
     Parse yaml_text as YAML.
     If parsing fails, raise error "malformed YAML".

  7. Extract fields from the parsed YAML into a Frontmatter record.
     Silently ignore any keys not listed below.

     depends_on:
       If present, read as a list of strings.
       If absent, use an empty list.

     external:
       If present, read as a list of external entries.
       For each entry:
         - "path" is required. If missing, raise error "malformed YAML".
         - "fragments" is optional. If present, read as a list of fragment entries.
           For each fragment entry:
             - "lines" is required. If missing, raise error "malformed YAML".
             - "hash" is required. If missing, raise error "malformed YAML".
             - "description" is optional.
           Build a FrontmatterExternalFragment record from these fields.
         Build a FrontmatterExternal record from path and fragments.
       If absent, use an empty list.

     input:
       If present, read as a string.
       If absent, use an empty string.

     outputs:
       If present, read as a list of output entries.
       For each entry:
         - "id" is required. If missing, raise error "malformed YAML".
         - "path" is required. If missing, raise error "malformed YAML".
         Build a FrontmatterOutput record from id and path.
       If absent, use an empty list.

  8. Return the populated Frontmatter record.
