<!-- code-from-spec: ROOT/functional/utils/frontmatter@TTHlJldRs88SO1lLcFeOV0IfITI -->

# ParseFrontmatter

Parses the optional YAML frontmatter block at the top of a spec file.
The frontmatter is delimited by `---` lines. If no frontmatter is
present, an empty record is returned. The file body is never read.

---

## Data Records

```
record ExternalFragment
  description: optional string   -- human-readable label for the fragment
  lines: string                  -- line range or selector (e.g. "1-10")
  hash: string                   -- content hash of the fragment

record External
  path: string                           -- path to the external file
  fragments: optional list of ExternalFragment  -- selected fragments within the file

record Output
  id: string    -- logical identifier for the output artifact
  path: string  -- relative file path where the artifact is written

record Frontmatter
  depends_on: list of strings   -- logical names of nodes this node depends on
  external: list of External    -- external file references
  input: string                 -- logical name of the input artifact node
  outputs: list of Output       -- artifacts this node is responsible for generating
```

All fields default to empty (empty list or empty string) when absent from the YAML.

---

## Functions

### ParseFrontmatter(file_path) -> Frontmatter

  Errors:
  - "file unreadable": the file at file_path cannot be opened or read.
  - "malformed YAML": the content between the `---` delimiters is not valid YAML.

  Steps:

  1. Open the file at file_path using OpenFileReader.
     If the file cannot be opened, raise error "file unreadable".

  2. Read the first line using ReadLine.
     If "end of file" is raised, close the reader and return an empty Frontmatter record.

  3. If the first line is not exactly "---":
     Close the reader and return an empty Frontmatter record.
     This is not an error — frontmatter is optional.

  4. Collect lines for the YAML body:
     Initialize an empty list called yaml_lines.

     For each subsequent line read with ReadLine:
       If "end of file" is raised before a closing "---" is found:
         Close the reader and return an empty Frontmatter record.
         (A file that starts with "---" but has no closing delimiter
          is treated as having no frontmatter.)

       If the line is exactly "---":
         Stop collecting. The frontmatter block is complete.

       Otherwise:
         Append the line to yaml_lines.

  5. Close the reader.
     (The file body after the closing "---" is never read.)

  6. If yaml_lines is empty:
     Return an empty Frontmatter record.
     An empty block (--- followed immediately by ---) is valid and produces empty fields.

  7. Join yaml_lines with newline characters to form a single YAML string.
     Parse that string as YAML.
     If parsing fails, raise error "malformed YAML".

  8. Extract the following fields from the parsed YAML.
     Ignore any fields not listed here (unknown fields are silently discarded).

     - "depends_on":
         If present, read as a list of strings.
         If absent, use an empty list.

     - "external":
         If present, read as a list of External records.
         For each entry in the list:
           - "path": string (required field within each entry)
           - "fragments": optional list of ExternalFragment records.
               For each fragment entry:
                 - "description": optional string
                 - "lines": string
                 - "hash": string
         If absent, use an empty list.

     - "input":
         If present, read as a string.
         If absent, use an empty string.

     - "outputs":
         If present, read as a list of Output records.
         For each entry in the list:
           - "id": string
           - "path": string
         If absent, use an empty list.

  9. Return the populated Frontmatter record.
```
