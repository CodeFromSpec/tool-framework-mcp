<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@Y0VFPHy4f98dzStomPukGYRwSa8 -->

# Frontmatter Parse — Test Specification

---

## Happy Path

---

### Test: Parses complete frontmatter (all fields)

**Setup**

Create a file containing frontmatter with all fields:
- `depends_on`: a list of two dependency strings
- `external`: one entry with a `path` and one fragment containing `description`, `lines`, and `hash`
- `input`: a non-empty string value
- `outputs`: two entries, each with `id` and `path`

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `depends_on` contains exactly the two listed dependency strings.
- `external` has exactly one entry with the correct `path`.
  - That entry has one fragment with the correct `description`, `lines`, and `hash`.
- `input` matches the specified string value.
- `outputs` has exactly two entries, each with the correct `id` and `path`.

---

### Test: Parses frontmatter with only outputs

**Setup**

Create a file containing frontmatter with only the `outputs` field:
- `outputs`: one entry with an `id` and a `path`

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` has exactly one entry with the correct `id` and `path`.

---

### Test: Parses frontmatter with only depends_on

**Setup**

Create a file containing frontmatter with only the `depends_on` field:
- `depends_on`: a list of two strings

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `depends_on` contains exactly the two listed strings.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

### Test: Parses frontmatter with external and fragments

**Setup**

Create a file containing frontmatter with an `external` field containing two entries:
- First entry: has a `path` and two fragments, each with `description`, `lines`, and `hash`
- Second entry: has a `path` only, no `fragments`

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `external` has exactly two entries.
- First entry has the correct `path` and exactly two fragments, each with the correct `description`, `lines`, and `hash`.
- Second entry has the correct `path` and no fragments.

---

### Test: Parses frontmatter with input field

**Setup**

Create a file containing frontmatter with only the `input` field set to a non-empty string.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `input` matches the specified string value.
- `depends_on` is empty.
- `external` is empty.
- `outputs` is empty.

---

### Test: Fragment without description

**Setup**

Create a file containing frontmatter with one `external` entry. That entry has one fragment with `lines` and `hash` set, but no `description`.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `external` has one entry with one fragment.
- The fragment has `description` absent (empty / not present).
- The fragment's `lines` and `hash` match the specified values.

---

### Test: Ignores unknown frontmatter fields

**Setup**

Create a file containing frontmatter with known fields (`depends_on`, `outputs`) plus an extra unknown field (e.g., `custom_field: value`).

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- Known fields are parsed correctly.
- The unknown field is silently ignored — it does not appear in the result and does not cause an error.

---

### Test: File with no frontmatter returns empty result

**Setup**

Create a file containing only body content and no `---` delimiter anywhere.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

## Edge Cases

---

### Test: Empty frontmatter

**Setup**

Create a file whose frontmatter delimiters are present (opening `---` and closing `---`) but nothing appears between them.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty.
- `outputs` is empty.

---

### Test: File with only frontmatter, nothing after

**Setup**

Create a file that has frontmatter (with at least one field set) followed by the closing `---` and no body content after it.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- Fields present in the frontmatter are parsed correctly.

---

### Test: Delimiter with trailing whitespace is not recognized

**Setup**

Create a file whose first line is `---   ` (three dashes followed by trailing spaces). The rest of the file is body content with no other `---` line.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

No error.
- The first line is not recognized as a frontmatter delimiter.
- Result has all fields empty.

---

## Failure Cases

---

### Test: File does not exist

**Setup**

Prepare a `PathCfs` that points to a file that does not exist on disk.

**Action**

Call `FrontmatterParse` with that path.

**Expected outcome**

Raises error "file unreadable".

---

### Test: Propagates path errors

**Setup**

Prepare an invalid `PathCfs` that would cause a directory traversal (e.g., `"../../outside"`).

**Action**

Call `FrontmatterParse` with that path.

**Expected outcome**

Raises error "directory traversal" propagated from `FileOpen`.

---

### Test: Malformed YAML

**Setup**

Create a file where the content between the `---` delimiters is not valid YAML (e.g., inconsistent indentation, invalid characters, or broken structure).

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

Raises error "malformed YAML".

---

### Test: Unclosed frontmatter block

**Setup**

Create a file that begins with `---` but has no second `---` line anywhere in the file.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

Raises error "malformed YAML".

---

### Test: Missing required field in external entry

**Setup**

Create a file with an `external` entry that has no `path` field (only fragments or other attributes).

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

Raises error "malformed YAML".

---

### Test: Missing required field in fragment

**Setup**

Create a file with an `external` entry containing a fragment that has `lines` set but no `hash` field.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

Raises error "malformed YAML".

---

### Test: Missing required field in output entry

**Setup**

Create a file with an `outputs` entry that has `id` set but no `path` field.

**Action**

Call `FrontmatterParse` with the path to that file.

**Expected outcome**

Raises error "malformed YAML".
