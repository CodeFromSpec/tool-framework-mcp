<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@qLlGdRxV3BzC8czImm87FdEo6mo -->

# Test Specification: FrontmatterParse

## Data Records

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

---

## Happy Path

### Test: Parses complete frontmatter (all fields)

**Setup:**
Create a file whose frontmatter contains all four fields:
- `depends_on` — a list with at least two entries (e.g., `["dep-a", "dep-b"]`)
- `external` — one entry with a `path` value and one fragment containing
  `description`, `lines`, and `hash`
- `input` — a non-empty string (e.g., `"some/input/file.md"`)
- `outputs` — two entries, each with `id` and `path`

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `depends_on` contains exactly the listed dependency strings.
- `external` has one entry whose `path` matches the specified value.
  That entry has one fragment with `description`, `lines`, and `hash`
  all matching the specified values.
- `input` matches the specified string.
- `outputs` has two entries; each entry's `id` and `path` match the
  specified values.

---

### Test: Parses frontmatter with only outputs

**Setup:**
Create a file whose frontmatter contains only the `outputs` field with
one entry (e.g., `id: "out-1"`, `path: "some/path/file.go"`).

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` has one entry with the correct `id` and `path`.

---

### Test: Parses frontmatter with only depends_on

**Setup:**
Create a file whose frontmatter contains only `depends_on` with two or
more string values.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `depends_on` contains exactly the listed strings.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

### Test: Parses frontmatter with external and fragments

**Setup:**
Create a file whose frontmatter contains an `external` list with two
entries:
- First entry: has a `path` and two fragments, each with `description`,
  `lines`, and `hash`.
- Second entry: has a `path` only, no `fragments` field.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `external` has two entries.
- First entry's `path` matches. It has two fragments; each fragment's
  `description`, `lines`, and `hash` match the specified values.
- Second entry's `path` matches. Its `fragments` list is empty or absent.

---

### Test: Parses frontmatter with input field

**Setup:**
Create a file whose frontmatter contains only the `input` field with a
non-empty string value.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `input` matches the specified string.
- `depends_on` is empty.
- `external` is empty.
- `outputs` is empty.

---

### Test: Fragment without description

**Setup:**
Create a file whose frontmatter contains one `external` entry with one
fragment that has `lines` and `hash` but no `description` field.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- The fragment's `description` is absent (not set).
- The fragment's `lines` and `hash` match the specified values.

---

### Test: Ignores unknown frontmatter fields

**Setup:**
Create a file whose frontmatter contains valid known fields (e.g.,
`depends_on`) plus one or more extra unknown fields (e.g.,
`custom_field: value`, `another_field: 42`).

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- Known fields are parsed correctly.
- Unknown fields are silently ignored — they do not appear in the result.

---

### Test: File with no frontmatter returns empty result

**Setup:**
Create a file whose content contains no `---` delimiter at all — only
body text.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

## Edge Cases

### Test: Empty frontmatter

**Setup:**
Create a file that begins with `---` on the first line and has a second
line of `---` immediately after, with no content between them.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

### Test: File with only frontmatter, nothing after

**Setup:**
Create a file that has a valid frontmatter block (opening `---`,
some fields, closing `---`) and no body content after the closing
delimiter.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- Fields parsed correctly from the frontmatter.

---

### Test: Delimiter with trailing whitespace is not recognized

**Setup:**
Create a file whose first line is `---   ` (three dashes followed by
trailing spaces). The rest of the file is body content with no other
`---` line.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- No error.
- Result has all fields empty — the line with trailing spaces is not
  treated as a frontmatter opening delimiter.

---

## Failure Cases

### Test: File does not exist

**Setup:**
Prepare a `PathCfs` value pointing to a file that does not exist on
disk.

**Action:**
Call `FrontmatterParse` with that path.

**Expected outcome:**
- Error `FileUnreadable` is returned.

---

### Test: Propagates path errors

**Setup:**
Prepare an invalid `PathCfs` value that attempts directory traversal
(e.g., `"../../outside"`).

**Action:**
Call `FrontmatterParse` with that path.

**Expected outcome:**
- Error `DirectoryTraversal` is returned, propagated from
  `FileReader`/`PathUtils` via `FileOpen`.

---

### Test: Malformed YAML

**Setup:**
Create a file that has an opening `---` delimiter followed by content
that is not valid YAML (e.g., unbalanced brackets or invalid
indentation), then a closing `---`.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### Test: Unclosed frontmatter block

**Setup:**
Create a file that begins with `---` but has no second `---` line
anywhere in the file.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### Test: Missing required field in external entry

**Setup:**
Create a file whose frontmatter contains an `external` list with one
entry that has no `path` field (only optional or unrelated fields).

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### Test: Missing required field in fragment

**Setup:**
Create a file whose frontmatter contains an `external` entry with one
fragment that has `lines` but no `hash` field.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### Test: Missing required field in output entry

**Setup:**
Create a file whose frontmatter contains an `outputs` entry that has
`id` but no `path` field.

**Action:**
Call `FrontmatterParse` with the path to that file.

**Expected outcome:**
- Error `MalformedYAML` is returned.
