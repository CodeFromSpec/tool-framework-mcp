<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@aHDP0XkeMvEU5eg-VTP1lkiAAkY -->

## Happy path

### Parses complete frontmatter (all fields)

Setup: Create a file with all four fields in frontmatter:
- `depends_on`: list of two dependencies
- `external`: list of two entries each with a `path` field
- `input`: a non-empty string value
- `output`: a non-empty path string

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `depends_on` contains both listed dependencies.
`external` has two entries each with the correct `path`. `input` matches the
specified value. `output` matches the specified path.

---

### Parses frontmatter with only output

Setup: Create a file with only `output` in frontmatter.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `depends_on` is empty. `external` is empty.
`input` is empty. `output` matches the specified path.

---

### Parses frontmatter with only depends_on

Setup: Create a file with only `depends_on` listing two values in frontmatter.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `depends_on` contains both listed values.
`external` is empty. `input` is empty. `output` is empty.

---

### Parses frontmatter with external entries

Setup: Create a file with `external` containing two entries each with a `path` field.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `external` has two entries each with the correct `path`.
All other fields are empty.

---

### Parses frontmatter with input field

Setup: Create a file with only the `input` field in frontmatter.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `input` matches the specified value. `depends_on` is empty.
`external` is empty. `output` is empty.

---

### Ignores unknown frontmatter fields

Setup: Create a file with known fields (`output`, `depends_on`) plus an extra unknown
field (e.g., `custom_field: value`) in frontmatter.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. Known fields parsed correctly. Unknown fields are ignored.

---

### File with no frontmatter returns empty result

Setup: Create a file with body content only — no `---` delimiter present.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `depends_on` is empty. `external` is empty. `input` is empty.
`output` is empty.

---

## Edge cases

### Empty frontmatter

Setup: Create a file with an opening `---` and closing `---` with nothing between them.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. `depends_on` is empty. `external` is empty. `input` is empty.
`output` is empty.

---

### File with only frontmatter, nothing after

Setup: Create a file with frontmatter fields followed by the closing `---` and no body content.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. Fields parsed correctly from frontmatter.

---

### Delimiter with trailing whitespace is not recognized

Setup: Create a file where the very first line is `---   ` (with trailing spaces).

Action: Call `FrontmatterParse` with the file path.

Expected outcome: No error. Result has all fields empty — the line with trailing spaces is
not recognized as a frontmatter delimiter.

---

## Failure cases

### File does not exist

Setup: No file is created.

Action: Call `FrontmatterParse` with a `pathutils.PathCfs` pointing to a non-existent file.

Expected outcome: Error FileUnreadable.

---

### Propagates path errors

Setup: No file is created.

Action: Call `FrontmatterParse` with an invalid `pathutils.PathCfs` (e.g., `"../../outside"`).

Expected outcome: Error DirectoryTraversal propagated from FileReader/PathUtils via FileOpen.

---

### Malformed YAML

Setup: Create a file with invalid YAML content between the frontmatter `---` delimiters.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: Error MalformedYAML.

---

### Unclosed frontmatter block

Setup: Create a file that begins with `---` but has no closing `---` — only body content follows.

Action: Call `FrontmatterParse` with the file path.

Expected outcome: Error MalformedYAML.

---

### Missing required field in external entry

Setup: Create a file with an `external` entry that has no `path` field (e.g., the entry is
an object with only an unrecognized key or is empty).

Action: Call `FrontmatterParse` with the file path.

Expected outcome: Error MalformedYAML.
