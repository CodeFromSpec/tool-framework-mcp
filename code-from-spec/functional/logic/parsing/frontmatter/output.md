<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@a9LpM9CPtZa-9VLC10jg1-O8cb4 -->

# Frontmatter Parser

namespace: frontmatter

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

All fields default to empty (empty list, empty string) when absent from the YAML.

## Functions

```
function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters
      is not valid YAML, or a required sub-record field
      is missing.
    - (FileReader.*): propagated from FileOpen.
```

### FrontmatterParse

1. Call `FileOpen(file_path)`.
   If `FileOpen` raises `FileUnreadable`, propagate as `FileUnreadable`.
   If `FileOpen` raises any `PathUtils.*` error, propagate it.
   Store the result as `reader`.

2. Call `FileReadLine(reader)` to read the first line.
   If `EndOfFile` is raised, call `FileClose(reader)` and return an empty Frontmatter record.
   Store the result as `first_line`.

3. If `first_line` is not exactly `"---"`:
   Call `FileClose(reader)` and return an empty Frontmatter record.

4. Collect YAML lines into an empty list called `yaml_lines`.
   Repeat:
     a. Call `FileReadLine(reader)`.
        If `EndOfFile` is raised:
          Call `FileClose(reader)`.
          Raise error `"malformed YAML"`.
     b. If the line is exactly `"---"`:
          Stop collecting. Exit the loop.
     c. Otherwise, append the line to `yaml_lines`.

5. Call `FileClose(reader)`.

6. Join `yaml_lines` with newline characters into a single string called `yaml_text`.

7. If `yaml_text` is empty or contains only whitespace:
   Return an empty Frontmatter record.

8. Parse `yaml_text` as YAML.
   If parsing fails, raise error `"malformed YAML"`.
   Store the parsed result as `parsed`.

9. Extract `depends_on` from `parsed`:
   If the key `"depends_on"` is present, read its value as a list of strings.
   Otherwise, use an empty list.

10. Extract `input` from `parsed`:
    If the key `"input"` is present, read its value as a string.
    Otherwise, use an empty string.

11. Extract `outputs` from `parsed`:
    If the key `"outputs"` is present, read its value as a list.
    Otherwise, use an empty list.
    For each entry in the list:
      a. If `"id"` is missing from the entry, raise error `"malformed YAML"`.
      b. If `"path"` is missing from the entry, raise error `"malformed YAML"`.
      c. Create a FrontmatterOutput record with:
           id: the value of `"id"`
           path: the value of `"path"`
    Collect results into `outputs_list`.

12. Extract `external` from `parsed`:
    If the key `"external"` is present, read its value as a list.
    Otherwise, use an empty list.
    For each entry in the list:
      a. If `"path"` is missing from the entry, raise error `"malformed YAML"`.
      b. Extract `fragments`:
           If `"fragments"` is present in the entry, read its value as a list.
           Otherwise, set fragments to absent (optional field not set).
           If fragments is present, for each fragment:
             i.  If `"lines"` is missing, raise error `"malformed YAML"`.
             ii. If `"hash"` is missing, raise error `"malformed YAML"`.
             iii.Create a FrontmatterExternalFragment record with:
                   description: value of `"description"` if present, otherwise absent
                   lines: value of `"lines"`
                   hash: value of `"hash"`
           Collect fragment records into `fragments_list`.
      c. Create a FrontmatterExternal record with:
           path: the value of `"path"`
           fragments: `fragments_list` if present, otherwise absent
    Collect results into `external_list`.

13. Return a Frontmatter record with:
      depends_on: `depends_on` list from step 9
      external: `external_list` from step 12
      input: `input` string from step 10
      outputs: `outputs_list` from step 11
```
