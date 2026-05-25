<!-- code-from-spec: ROOT/functional/utils/frontmatter@PENDING -->

record ExternalFragment
  description: optional string
  lines: string
  hash: string

record External
  path: string
  fragments: optional list of ExternalFragment

record Output
  id: string
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of External
  input: string
  outputs: list of Output

function ParseFrontmatter(file_path) -> frontmatter

  1. Open the file using OpenFileReader(file_path).
     If the file cannot be opened, raise error "file unreadable".

  2. Read the first line using ReadLine.
     If end of file is reached, return an empty Frontmatter record.

  3. If the first line is not exactly "---", return an empty
     Frontmatter record.

  4. Collect lines into a list called yaml_lines:
     for each subsequent line from ReadLine:
       if the line is exactly "---", stop collecting.
       otherwise, append the line to yaml_lines.
     If end of file is reached before a closing "---",
     raise error "malformed YAML".

  5. Join yaml_lines with LF into a single string.
     Parse the string as YAML.
     If parsing fails, raise error "malformed YAML".

  6. Extract known fields from the parsed YAML into a
     Frontmatter record:
     - depends_on: if present, a list of strings. Otherwise empty list.
     - external: if present, a list of External records.
       For each entry, extract path (string) and optionally
       fragments (list of ExternalFragment records, each with
       optional description, lines, and hash).
       Otherwise empty list.
     - input: if present, a string. Otherwise empty string.
     - outputs: if present, a list of Output records, each
       with id (string) and path (string).
       Otherwise empty list.

  7. Ignore any fields not listed above.

  8. Return the Frontmatter record.
