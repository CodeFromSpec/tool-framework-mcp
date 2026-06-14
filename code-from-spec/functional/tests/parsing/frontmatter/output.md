<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@VoDHN1kK5WgdxZpoKHBzLo7l8DA -->

## Test Suite: FrontmatterParse

---

### TC-01: Parses complete frontmatter (all fields)

Setup:
  Create a file with frontmatter containing:
    - `depends_on` with entries: one `SPEC/` prefixed, one `ARTIFACT/` prefixed, one `EXTERNAL/` prefixed
    - `input` set to a non-empty string
    - `output` set to a non-empty path string

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` contains exactly the three listed dependency entries.
  `input` matches the specified value.
  `output` matches the specified path.

---

### TC-02: Parses frontmatter with only output

Setup:
  Create a file with frontmatter containing only `output` set to a path string.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` is empty.
  `input` is empty.
  `output` matches the specified path.

---

### TC-03: Parses frontmatter with only depends_on

Setup:
  Create a file with frontmatter containing only `depends_on` with two or more entries.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` contains the listed values.
  `input` is empty.
  `output` is empty.

---

### TC-04: Parses frontmatter with EXTERNAL/ in depends_on

Setup:
  Create a file with frontmatter containing `depends_on` with the entry `EXTERNAL/proto/api.proto`.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` contains `EXTERNAL/proto/api.proto`.

---

### TC-05: Parses frontmatter with input field only

Setup:
  Create a file with frontmatter containing only `input` set to a non-empty string.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `input` matches the specified value.
  `depends_on` is empty.
  `output` is empty.

---

### TC-06: Ignores unknown frontmatter fields

Setup:
  Create a file with frontmatter containing all known fields (`depends_on`, `input`, `output`)
  plus an additional unknown field (e.g., `custom_field: value`).

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  Known fields are parsed correctly.
  The unknown field is silently ignored.

---

### TC-07: File with no frontmatter returns empty result

Setup:
  Create a file containing only body content with no `---` delimiter anywhere.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` is empty.
  `input` is empty.
  `output` is empty.

---

### TC-08: Empty frontmatter

Setup:
  Create a file with an opening `---` line, a closing `---` line, and nothing between them.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  `depends_on` is empty.
  `input` is empty.
  `output` is empty.

---

### TC-09: File with only frontmatter, nothing after

Setup:
  Create a file with valid frontmatter followed by a closing `---` and no body content after.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  Fields present in the frontmatter are parsed correctly.

---

### TC-10: Delimiter with trailing whitespace is not recognized

Setup:
  Create a file where the very first line is `---   ` (three dashes followed by trailing spaces).

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  Result has all fields empty — the line with trailing whitespace is not treated as a frontmatter delimiter.

---

### TC-11: File does not exist

Setup:
  A `PathCfs` pointing to a path where no file exists.

Action:
  Call `FrontmatterParse` with that path.

Expected outcome:
  Error FileUnreadable.

---

### TC-12: Propagates path errors

Setup:
  An invalid `PathCfs` value such as `"../../outside"` that would trigger a traversal violation.

Action:
  Call `FrontmatterParse` with that path.

Expected outcome:
  Error DirectoryTraversal (propagated from FileReader/PathUtils via FileOpen).

---

### TC-13: Malformed YAML

Setup:
  Create a file with an opening `---`, invalid YAML content between the delimiters, and a closing `---`.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  Error MalformedYAML.

---

### TC-14: Unclosed frontmatter block

Setup:
  Create a file that starts with `---` followed by valid YAML-like content but no closing `---`.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  Error MalformedYAML.

---

### TC-15: Unknown field 'external' is silently ignored

Setup:
  Create a file with frontmatter containing an `external` field (v3 legacy format) in addition to
  any known fields.

Action:
  Call `FrontmatterParse` with the file path.

Expected outcome:
  No error.
  The `external` field is ignored.
  Only `depends_on`, `input`, and `output` are recognized and returned.
