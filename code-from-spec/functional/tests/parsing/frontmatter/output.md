<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@cNudHADxoatpkjMk3D_dsFetLj8 -->

## Test Cases: FrontmatterParse

---

### Happy Path

---

#### Parses complete frontmatter (all fields)

Setup: Create a file with frontmatter containing all fields:
- `depends_on`: two dependency strings
- `external`: two entries each with a `path` field
- `input`: a non-empty string value
- `outputs`: two entries each with `id` and `path` fields

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- `depends_on` contains the two listed dependency strings.
- `external` has two entries with the correct `path` values.
- `input` matches the specified value.
- `outputs` has two entries with the correct `id` and `path` values.

---

#### Parses frontmatter with only outputs

Setup: Create a file with frontmatter containing only an `outputs` field with one entry having `id` and `path`.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` has one entry with the correct `id` and `path`.

---

#### Parses frontmatter with only depends_on

Setup: Create a file with frontmatter containing only a `depends_on` field listing two values.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- `depends_on` contains the two listed values.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

#### Parses frontmatter with external entries

Setup: Create a file with frontmatter containing only an `external` field with two entries each having a `path`.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- `external` has two entries with the correct `path` values.
- All other fields are empty.

---

#### Parses frontmatter with input field

Setup: Create a file with frontmatter containing only an `input` field set to a non-empty string.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- `input` matches the specified value.
- `depends_on` is empty.
- `external` is empty.
- `outputs` is empty.

---

#### Ignores unknown frontmatter fields

Setup: Create a file with frontmatter containing known fields (`depends_on`, `outputs`) plus an extra unknown field (e.g., `custom_field: value`).

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- Known fields are parsed correctly.
- Unknown fields are silently ignored.

---

#### File with no frontmatter returns empty result

Setup: Create a file containing only body content, with no `---` delimiter anywhere.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- All fields (`depends_on`, `external`, `input`, `outputs`) are empty.

---

### Edge Cases

---

#### Empty frontmatter

Setup: Create a file with an opening `---` and a closing `---` and nothing between them.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- All fields are empty.

---

#### File with only frontmatter, nothing after

Setup: Create a file with valid frontmatter (opening `---`, some fields, closing `---`) and no body content after the closing delimiter.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- Fields are parsed correctly.

---

#### Delimiter with trailing whitespace is not recognized

Setup: Create a file where the first line is `---   ` (with trailing spaces), followed by some content.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- No error.
- All fields are empty — the line with trailing spaces is not recognized as a frontmatter delimiter.

---

### Failure Cases

---

#### File does not exist

Setup: No file is created. Prepare a `PathCfs` pointing to a non-existent file.

Action: Call `FrontmatterParse` with the non-existent file path.

Expected outcome:
- Error `FileUnreadable`.

---

#### Propagates path errors

Setup: Prepare an invalid `PathCfs` value that represents a path escaping the root (e.g., `"../../outside"`).

Action: Call `FrontmatterParse` with the invalid path.

Expected outcome:
- Error `DirectoryTraversal` propagated from `FileReader`/`PathUtils` via `FileOpen`.

---

#### Malformed YAML

Setup: Create a file with a valid opening `---`, invalid YAML content between the delimiters (e.g., unbalanced brackets or bad indentation), and a closing `---`.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- Error `MalformedYAML`.

---

#### Unclosed frontmatter block

Setup: Create a file that begins with `---` but has no subsequent closing `---` — only body content follows.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- Error `MalformedYAML`.

---

#### Missing required field in external entry

Setup: Create a file with an `external` entry in the frontmatter that has no `path` field (e.g., the entry is present but `path` is omitted).

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- Error `MalformedYAML`.

---

#### Missing required field in output entry

Setup: Create a file with an `outputs` entry that has an `id` field but no `path` field.

Action: Call `FrontmatterParse` with the file path.

Expected outcome:
- Error `MalformedYAML`.
