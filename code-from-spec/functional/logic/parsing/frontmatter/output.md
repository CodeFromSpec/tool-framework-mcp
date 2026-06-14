<!-- code-from-spec: ROOT/functional/logic/parsing/frontmatter@KD7r7zcMAC4P_rxLno1l5ABqt-k -->

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
      or an opening "---" was found but no closing "---" exists.
    - (FileReader.*): propagated from FileOpen.

  1. Call FileOpen(file_path) to obtain a reader.
     If FileOpen raises FileUnreadable or any FileReader error, propagate it.

  2. Read the first line using FileReadLine.
     If FileReadLine raises EndOfFile, call FileClose(reader) and
     return an empty Frontmatter record with:
       depends_on = empty list
       input = ""
       output = ""

  3. If the first line is not exactly "---", call FileClose(reader) and
     return an empty Frontmatter record with:
       depends_on = empty list
       input = ""
       output = ""

  4. Collect YAML lines:
     Initialize yaml_lines as an empty list.
     Loop:
       Read the next line using FileReadLine.
       If FileReadLine raises EndOfFile:
         Call FileClose(reader).
         Raise error "malformed YAML".
       If the line is exactly "---":
         Break out of the loop.
       Append the line to yaml_lines.

  5. Call FileClose(reader).

  6. If yaml_lines is empty, return an empty Frontmatter record with:
       depends_on = empty list
       input = ""
       output = ""

  7. Join yaml_lines with newline characters to form yaml_text.
     Parse yaml_text as YAML.
     If parsing fails, raise error "malformed YAML".

  8. Extract fields from the parsed YAML:
     - depends_on: list of strings — default to empty list if absent.
     - input: string — default to "" if absent.
     - output: string — default to "" if absent.
     Silently ignore all other YAML keys.

  9. Return the Frontmatter record with the extracted fields.
