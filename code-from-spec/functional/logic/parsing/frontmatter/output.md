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

## Functions

```
function FrontmatterParse(file_path: PathCfs) -> Frontmatter
  errors:
    - (path errors): propagated from FileOpen.
    - file unreadable: the file cannot be opened or read.
    - malformed YAML: the content between --- delimiters is not valid YAML,
      or required fields within sub-records are missing,
      or an opening "---" is present but no closing "---" is found.

  1. Open the file at file_path using FileOpen.
     If FileOpen raises an error, propagate it.

  2. Read the first line using FileReadLine.
     If FileReadLine raises "end of file", close the reader with FileClose
     and return an empty Frontmatter record.

  3. If the first line is not exactly "---", close the reader with FileClose
     and return an empty Frontmatter record.

  4. Collect YAML lines:
     Set yaml_lines to an empty list.
     Loop:
       Read the next line using FileReadLine.
       If FileReadLine raises "end of file",
         close the reader with FileClose
         and raise error "malformed YAML".
       If the line is exactly "---",
         stop collecting and proceed to step 5.
       Otherwise, append the line to yaml_lines.

  5. Close the reader with FileClose.

  6. Join yaml_lines with newline as a single string.
     Parse the joined string as YAML.
     If parsing fails, raise error "malformed YAML".

  7. If the parsed YAML is empty or not a mapping,
     return an empty Frontmatter record.

  8. Extract fields from the parsed YAML mapping.
     Unknown keys are silently ignored.

     Extract "depends_on":
       If present, read as a list of strings.
       If absent, use an empty list.

     Extract "external":
       If present, read as a list of external entries.
       For each entry:
         If "path" is missing or empty,
           raise error "malformed YAML".
         Set external_path to the value of "path".
         If "fragments" is present, process each fragment:
           If "lines" is missing or empty,
             raise error "malformed YAML".
           If "hash" is missing or empty,
             raise error "malformed YAML".
           Set fragment_description to the value of "description"
             if present, otherwise absent.
           Construct a FrontmatterExternalFragment record with
             description: fragment_description,
             lines: value of "lines",
             hash: value of "hash".
         Construct a FrontmatterExternal record with
           path: external_path,
           fragments: the processed fragment list if "fragments" was present,
             otherwise absent.
       If absent, use an empty list.

     Extract "input":
       If present, read as a string.
       If absent, use an empty string.

     Extract "outputs":
       If present, read as a list of output entries.
       For each entry:
         If "id" is missing or empty,
           raise error "malformed YAML".
         If "path" is missing or empty,
           raise error "malformed YAML".
         Construct a FrontmatterOutput record with
           id: value of "id",
           path: value of "path".
       If absent, use an empty list.

  9. Construct and return a Frontmatter record with
       depends_on: extracted depends_on list,
       external: extracted external list,
       input: extracted input string,
       outputs: extracted outputs list.
```
