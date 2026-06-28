<!-- code-from-spec: SPEC/functional/logic/parsing/frontmatter@KMaOqkGDYPRPy4UjLpllmjIt170 -->

namespace: frontmatter

---

record Frontmatter
  depends_on: list of strings
  input: string
  output: string

---

function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters is not valid YAML,
      or an opening --- is found but no closing --- follows.
    - (FileReader.*): propagated from FileOpen.

  1. Call FileOpen(file_path, "read", 30000).
     If FileOpen raises FileUnreadable or any propagated error, re-raise as FileUnreadable.

  2. Call FileReadLine to read the first line.
     If EndOfFile is raised, call FileClose and return an empty Frontmatter record
       with depends_on = [], input = "", output = "".
     If the first line is not exactly "---", call FileClose and return an empty
       Frontmatter record with depends_on = [], input = "", output = "".

  3. Collect YAML lines:
     Initialize yaml_lines as an empty list of strings.
     Repeat:
       Call FileReadLine.
       If EndOfFile is raised:
         Call FileClose.
         Raise error "malformed YAML".
       If the line is exactly "---":
         Stop collecting.
       Else:
         Append the line to yaml_lines.

  4. Call FileClose.

  5. If yaml_lines is empty, return an empty Frontmatter record
     with depends_on = [], input = "", output = "".

  6. Join yaml_lines into a single string, each line separated by a newline character.
     Parse the joined string as YAML.
     If parsing fails, raise error "malformed YAML".

  7. From the parsed YAML, extract the following fields, ignoring all other keys:
     - depends_on: list of strings. If absent or null, use [].
     - input: string. If absent or null, use "".
     - output: string. If absent or null, use "".

  8. Return a Frontmatter record with the extracted field values.
