<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@gDq-HjK3wiB6rvbrungalv2tky0 -->

# FrontmatterParse

## Data Structures

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

---

## Functions

### FrontmatterParse(file_path: PathCfs) -> Frontmatter

  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters is not valid YAML,
      or a required field in a sub-record is missing.
    - (FileReader.*): propagated from FileOpen.

  1. Open the file at <file_path> using FileOpen.
     If FileOpen raises an error, propagate it.
     If the file cannot be read, raise error "FileUnreadable".

  2. Read the first line using FileReadLine.
     If EndOfFile is raised, close the reader and return an empty Frontmatter record.
     If the first line is not exactly "---", close the reader and return an empty Frontmatter record.

  3. Collect lines into a list called <yaml_lines>, starting empty.
     Repeat:
       a. Read the next line using FileReadLine.
          If EndOfFile is raised, close the reader and raise error "MalformedYAML"
          (opening "---" was found but closing "---" was not).
       b. If the line is exactly "---", stop collecting.
       c. Otherwise, append the line to <yaml_lines>.

  4. Close the reader using FileClose.

  5. Join <yaml_lines> into a single string <yaml_text> using newline as separator.

  6. Parse <yaml_text> as YAML into a raw mapping <raw>.
     If parsing fails, raise error "MalformedYAML".
     If <yaml_text> is empty or produces no mapping, proceed with an empty mapping.

  7. Build and return a Frontmatter record by extracting known fields from <raw>:

     a. depends_on:
        If the "depends_on" key is present, extract as a list of strings.
        Otherwise, use an empty list.

     b. external:
        If the "external" key is present, extract as a list.
        For each entry in the list:
          - If "path" is missing or empty, raise error "MalformedYAML".
          - Extract "path" as a string.
          - If "fragments" is present, extract as a list.
            For each fragment entry:
              - If "lines" is missing or empty, raise error "MalformedYAML".
              - If "hash" is missing or empty, raise error "MalformedYAML".
              - Extract "description" as an optional string (absent if not present).
              - Extract "lines" as a string.
              - Extract "hash" as a string.
              - Produce a FrontmatterExternalFragment record.
            Set fragments to the resulting list.
          - Otherwise, fragments is absent.
          - Produce a FrontmatterExternal record.
        Otherwise, use an empty list.

     c. input:
        If the "input" key is present, extract as a string.
        Otherwise, use an empty string.

     d. outputs:
        If the "outputs" key is present, extract as a list.
        For each entry in the list:
          - If "id" is missing or empty, raise error "MalformedYAML".
          - If "path" is missing or empty, raise error "MalformedYAML".
          - Extract "id" as a string.
          - Extract "path" as a string.
          - Produce a FrontmatterOutput record.
        Otherwise, use an empty list.

     Ignore any other keys present in <raw>.

  8. Return the Frontmatter record.
```

---

## Contracts

- The parser reads only the frontmatter block. It never reads the file body beyond
  the closing "---" delimiter.
- Unknown YAML keys at any level are silently ignored.
- All recognized fields are optional. An empty frontmatter block ("---" followed
  immediately by "---") produces a Frontmatter record with all fields at their
  defaults.
- If the first line is not exactly "---" (no leading or trailing whitespace
  permitted), the file is treated as having no frontmatter and an empty Frontmatter
  record is returned without error.
