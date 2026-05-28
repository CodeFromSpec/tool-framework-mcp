<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@GvZ7Gw3eUlA0xCUdnyCASSBKSzs -->

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

## Functions

```
function FrontmatterParse(file_path: PathCfs) -> Frontmatter
  errors:
    - (path errors): propagated from FileOpen
    - "file unreadable": the file cannot be opened or read
    - "malformed YAML": the content between --- delimiters is not valid YAML,
                        or a required field in a sub-record is missing

  1. Open the file at file_path using FileOpen.
     If FileOpen raises an error, propagate it to the caller.

  2. Read the first line using FileReadLine.
     If FileReadLine raises "end of file", close the reader with FileClose
     and return an empty Frontmatter record.

  3. If the first line is not exactly "---", close the reader with FileClose
     and return an empty Frontmatter record.

  4. Collect YAML lines:
     Set yaml_lines to an empty list.
     Set closing_delimiter_found to false.

     Repeat:
       Read the next line using FileReadLine.
       If FileReadLine raises "end of file":
         Close the reader with FileClose.
         Raise error "malformed YAML".
       If the line is exactly "---":
         Set closing_delimiter_found to true.
         Break out of the loop.
       Append the line to yaml_lines.

  5. Close the reader with FileClose.

  6. If yaml_lines is empty, return an empty Frontmatter record.

  7. Join yaml_lines with newline characters to form yaml_text.
     Parse yaml_text as YAML.
     If parsing fails, raise error "malformed YAML".

  8. Build the Frontmatter record from the parsed YAML:

     depends_on:
       If the YAML contains a "depends_on" key, read its value as a list
       of strings. If the key is absent, use an empty list.

     external:
       If the YAML contains an "external" key, read its value as a list.
       For each entry in the list:
         If "path" is missing or empty, raise error "malformed YAML".
         Set ext_path to the value of "path".
         If "fragments" is present:
           For each fragment entry:
             If "lines" is missing, raise error "malformed YAML".
             If "hash" is missing, raise error "malformed YAML".
             Build a FrontmatterExternalFragment record:
               description: value of "description" if present, otherwise absent
               lines: value of "lines"
               hash: value of "hash"
           Set fragments to the list of built FrontmatterExternalFragment records.
         Else:
           Set fragments to absent.
         Build a FrontmatterExternal record:
           path: ext_path
           fragments: fragments
       If the key is absent, use an empty list.

     input:
       If the YAML contains an "input" key, read its value as a string.
       If the key is absent, use an empty string.

     outputs:
       If the YAML contains an "outputs" key, read its value as a list.
       For each entry in the list:
         If "id" is missing, raise error "malformed YAML".
         If "path" is missing, raise error "malformed YAML".
         Build a FrontmatterOutput record:
           id: value of "id"
           path: value of "path"
       If the key is absent, use an empty list.

     Silently ignore any YAML keys not listed above.

  9. Return the Frontmatter record.
```
