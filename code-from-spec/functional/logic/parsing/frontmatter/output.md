<!-- code-from-spec: SPEC/functional/logic/parsing/frontmatter@SEPK5Dia33SWXJSoowYf-zSTy6o -->

namespace: frontmatter

record Frontmatter
  depends_on: list of strings
  input: string
  output: string


function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters is not valid YAML,
      or an opening --- is found but no closing --- follows.
    - (FileReader.*): propagated from FileOpen.

  1. Call FileOpen(file_path) to obtain a reader.
     If FileOpen raises FileUnreadable or any PathUtils error, propagate it.

  2. Call FileReadLine(reader) to read the first line.
     If it raises EndOfFile, call FileClose(reader) and return an empty
     Frontmatter record with depends_on = [], input = "", output = "".
     If the line is not exactly "---", call FileClose(reader) and return
     the same empty Frontmatter record.

  3. Collect YAML lines into a list, starting empty.
     Repeat:
       a. Call FileReadLine(reader).
          If it raises EndOfFile, call FileClose(reader) and
          raise error "malformed YAML".
       b. If the line is exactly "---", stop collecting.
       c. Otherwise append the line to the YAML lines list and continue.

  4. Call FileClose(reader).

  5. If the YAML lines list is empty, return an empty Frontmatter record
     with depends_on = [], input = "", output = "".

  6. Join the collected YAML lines with newline characters into a single
     string. Parse the result as YAML.
     If parsing fails, raise error "malformed YAML".

  7. From the parsed YAML mapping, extract the following fields,
     defaulting each to empty when absent:
     - depends_on: list of strings — default []
     - input: string — default ""
     - output: string — default ""
     Silently ignore any other keys present in the mapping.

  8. Return the populated Frontmatter record.
