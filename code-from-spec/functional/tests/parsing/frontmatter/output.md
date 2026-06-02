<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@WU7UgYJbHWMcsq0D0NbEOl2jWW0 -->

# Test Specification: FrontmatterParse

## Happy Path

### Parses complete frontmatter (all fields)

Setup: Create a file with the following content:
```
---
depends_on:
  - ROOT/a
  - ROOT/b
external:
  - path: some/external/file.md
  - path: another/external/file.md
input: path/to/input.md
output: path/to/output.md
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`depends_on` = `["ROOT/a", "ROOT/b"]`,
`external` has two `FrontmatterExternal` entries with `path` = `"some/external/file.md"` and `path` = `"another/external/file.md"`,
`input` = `"path/to/input.md"`,
`output` = `"path/to/output.md"`. No error.

---

### Parses frontmatter with only output

Setup: Create a file with the following content:
```
---
output: path/to/output.md
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`depends_on` = empty list,
`external` = empty list,
`input` = empty string,
`output` = `"path/to/output.md"`. No error.

---

### Parses frontmatter with only depends_on

Setup: Create a file with the following content:
```
---
depends_on:
  - ROOT/x
  - ROOT/y
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`depends_on` = `["ROOT/x", "ROOT/y"]`,
`external` = empty list,
`input` = empty string,
`output` = empty string. No error.

---

### Parses frontmatter with external entries

Setup: Create a file with the following content:
```
---
external:
  - path: docs/reference.md
  - path: docs/guide.md
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`external` having two entries: first with `path` = `"docs/reference.md"`, second with `path` = `"docs/guide.md"`. All other fields empty. No error.

---

### Parses frontmatter with input field

Setup: Create a file with the following content:
```
---
input: path/to/input.md
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`input` = `"path/to/input.md"`. All other fields empty. No error.

---

### Ignores unknown frontmatter fields

Setup: Create a file with the following content:
```
---
output: path/to/output.md
custom_field: some value
another_unknown: 42
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with
`output` = `"path/to/output.md"`. All other known fields empty. Unknown fields ignored. No error.

---

### File with no frontmatter returns empty result

Setup: Create a file with no `---` delimiter — body content only:
```
This is just body content.
No frontmatter here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with all fields empty. No error.

---

## Edge Cases

### Empty frontmatter

Setup: Create a file with the following content:
```
---
---
Body content here.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with all fields empty. No error.

---

### File with only frontmatter, nothing after

Setup: Create a file with the following content (no body after closing delimiter):
```
---
output: path/to/output.md
---
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with `output` = `"path/to/output.md"`. All other fields empty. No error.

---

### Delimiter with trailing whitespace is not recognized

Setup: Create a file where the first line is `---   ` (with trailing spaces):
```
---   
output: path/to/output.md
---
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Returns a `Frontmatter` record with all fields empty — the first line is not recognized as a delimiter, so no frontmatter is parsed. No error.

---

## Failure Cases

### File does not exist

Setup: None.

Action: Call `FrontmatterParse` with a `PathCfs` pointing to a non-existent file.

Expected outcome: Error `FileUnreadable`.

---

### Propagates path errors

Setup: None.

Action: Call `FrontmatterParse` with an invalid `PathCfs` such as `"../../outside"`.

Expected outcome: Error `DirectoryTraversal` propagated from `FileReader`/`PathUtils` via `FileOpen`.

---

### Malformed YAML

Setup: Create a file with invalid YAML between frontmatter delimiters:
```
---
key: [unclosed bracket
another: : invalid
---
Body content.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Error `MalformedYAML`.

---

### Unclosed frontmatter block

Setup: Create a file that starts with `---` but has no closing `---`:
```
---
output: path/to/output.md
Body content with no closing delimiter.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Error `MalformedYAML`.

---

### Missing required field in external entry

Setup: Create a file with an `external` entry that has no `path` field:
```
---
external:
  - name: some-name
---
Body content.
```

Action: Call `FrontmatterParse` with the path to that file.

Expected outcome: Error `MalformedYAML`.
