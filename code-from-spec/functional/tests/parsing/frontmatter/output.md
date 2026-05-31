<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@cl1UQur34DKetda98Rkp0RdLehM -->

# Test Specification: FrontmatterParse

## Records

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

### TC-HP-01: Parses complete frontmatter (all fields)

**Setup:**
Create a file with the following frontmatter containing all fields:
- `depends_on`: a list of two dependency strings
- `external`: one entry with a `path` and one fragment containing `description`, `lines`, and `hash`
- `input`: a non-empty string value
- `outputs`: two entries, each with `id` and `path`

Example file content:
```
---
depends_on:
  - "dep/one"
  - "dep/two"
external:
  - path: "some/external/file.md"
    fragments:
      - description: "A fragment description"
        lines: "10-20"
        hash: "abc123"
input: "some/input/file.md"
outputs:
  - id: "output_one"
    path: "path/to/output_one.go"
  - id: "output_two"
    path: "path/to/output_two.go"
---
Body content here.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` contains `"dep/one"` and `"dep/two"`.
- `external` has one entry with `path` equal to `"some/external/file.md"` and one fragment with `description` equal to `"A fragment description"`, `lines` equal to `"10-20"`, `hash` equal to `"abc123"`.
- `input` equals `"some/input/file.md"`.
- `outputs` has two entries: first with `id` `"output_one"` and `path` `"path/to/output_one.go"`, second with `id` `"output_two"` and `path` `"path/to/output_two.go"`.

---

### TC-HP-02: Parses frontmatter with only outputs

**Setup:**
Create a file with frontmatter containing only `outputs` (one entry with `id` and `path`).

Example file content:
```
---
outputs:
  - id: "my_output"
    path: "path/to/my_output.go"
---
Body content here.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty string.
- `outputs` has one entry with `id` `"my_output"` and `path` `"path/to/my_output.go"`.

---

### TC-HP-03: Parses frontmatter with only depends_on

**Setup:**
Create a file with frontmatter containing only `depends_on` (a list of two strings).

Example file content:
```
---
depends_on:
  - "dep/alpha"
  - "dep/beta"
---
Body content here.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` contains `"dep/alpha"` and `"dep/beta"`.
- `external` is empty.
- `input` is empty string.
- `outputs` is empty.

---

### TC-HP-04: Parses frontmatter with external and fragments

**Setup:**
Create a file with frontmatter containing an `external` field with two entries. The first entry has a `path` and two fragments each with `description`, `lines`, and `hash`. The second entry has only a `path` and no fragments.

Example file content:
```
---
external:
  - path: "docs/first.md"
    fragments:
      - description: "First fragment"
        lines: "1-5"
        hash: "hash001"
      - description: "Second fragment"
        lines: "10-15"
        hash: "hash002"
  - path: "docs/second.md"
---
Body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `external` has two entries.
- First entry: `path` equals `"docs/first.md"`, two fragments — first with `description` `"First fragment"`, `lines` `"1-5"`, `hash` `"hash001"`; second with `description` `"Second fragment"`, `lines` `"10-15"`, `hash` `"hash002"`.
- Second entry: `path` equals `"docs/second.md"`, no fragments (empty or absent).
- `depends_on` is empty, `input` is empty string, `outputs` is empty.

---

### TC-HP-05: Parses frontmatter with input field

**Setup:**
Create a file with frontmatter containing only the `input` field.

Example file content:
```
---
input: "path/to/input/source.md"
---
Body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `input` equals `"path/to/input/source.md"`.
- `depends_on` is empty.
- `external` is empty.
- `outputs` is empty.

---

### TC-HP-06: Fragment without description

**Setup:**
Create a file with frontmatter containing an `external` entry that has one fragment with `lines` and `hash` but no `description`.

Example file content:
```
---
external:
  - path: "docs/nodesc.md"
    fragments:
      - lines: "5-10"
        hash: "hashXYZ"
---
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `external` has one entry with `path` `"docs/nodesc.md"` and one fragment.
- The fragment has `description` absent (not present), `lines` equal to `"5-10"`, and `hash` equal to `"hashXYZ"`.

---

### TC-HP-07: Ignores unknown frontmatter fields

**Setup:**
Create a file with frontmatter containing known fields (`depends_on`, `outputs`) plus one or more unknown fields (e.g., `custom_field: "some value"`, `another_unknown: 42`).

Example file content:
```
---
depends_on:
  - "dep/known"
custom_field: "some value"
another_unknown: 42
outputs:
  - id: "out1"
    path: "path/out1.go"
---
Body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` contains `"dep/known"`.
- `outputs` has one entry with `id` `"out1"` and `path` `"path/out1.go"`.
- Unknown fields are silently ignored.

---

### TC-HP-08: File with no frontmatter returns empty result

**Setup:**
Create a file that contains no `---` delimiter at all — body content only.

Example file content:
```
This is just body content.
No frontmatter here.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty string.
- `outputs` is empty.

---

## Edge Cases

### TC-EC-01: Empty frontmatter

**Setup:**
Create a file with opening and closing `---` delimiters and nothing between them.

Example file content:
```
---
---
Some body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` is empty.
- `external` is empty.
- `input` is empty string.
- `outputs` is empty.

---

### TC-EC-02: File with only frontmatter, nothing after

**Setup:**
Create a file with a frontmatter block (opening `---`, content, closing `---`) and no body content after the closing `---`.

Example file content:
```
---
depends_on:
  - "dep/only"
---
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- `depends_on` contains `"dep/only"`.
- `external` is empty, `input` is empty string, `outputs` is empty.

---

### TC-EC-03: Delimiter with trailing whitespace is not recognized

**Setup:**
Create a file where the first line is `---   ` (three dashes followed by trailing spaces). No other `---` delimiter is present.

Example file content:
```
---   
Some body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- No error.
- The line `---   ` is not recognized as a frontmatter delimiter.
- Result has all fields empty (treated as a file with no frontmatter).

---

## Failure Cases

### TC-FC-01: File does not exist

**Setup:**
No file is created. Use a `PathCfs` that points to a non-existent file path within the allowed directory.

**Action:**
Call `FrontmatterParse` with the non-existent path.

**Expected outcome:**
- Error `FileUnreadable` is returned.

---

### TC-FC-02: Propagates path errors

**Setup:**
No file is created. Use an invalid `PathCfs` that represents a directory traversal (e.g., `"../../outside"`).

**Action:**
Call `FrontmatterParse` with the invalid path.

**Expected outcome:**
- Error `DirectoryTraversal` is returned (propagated from FileReader/PathUtils via FileOpen).

---

### TC-FC-03: Malformed YAML

**Setup:**
Create a file with invalid YAML content between the frontmatter delimiters.

Example file content:
```
---
depends_on: [unclosed bracket
  - bad: yaml: here
---
Body content.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### TC-FC-04: Unclosed frontmatter block

**Setup:**
Create a file that starts with `---` but has no closing `---` delimiter.

Example file content:
```
---
depends_on:
  - "dep/one"
Body content with no closing delimiter.
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### TC-FC-05: Missing required field in external entry

**Setup:**
Create a file with an `external` entry that has no `path` field (only fragments or other fields).

Example file content:
```
---
external:
  - fragments:
      - lines: "1-5"
        hash: "abc"
---
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### TC-FC-06: Missing required field in fragment

**Setup:**
Create a file with an `external` entry containing a fragment that has `lines` but no `hash`.

Example file content:
```
---
external:
  - path: "docs/file.md"
    fragments:
      - description: "Some fragment"
        lines: "1-10"
---
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- Error `MalformedYAML` is returned.

---

### TC-FC-07: Missing required field in output entry

**Setup:**
Create a file with an `outputs` entry that has `id` but no `path`.

Example file content:
```
---
outputs:
  - id: "output_without_path"
---
```

**Action:**
Call `FrontmatterParse` with the path to this file.

**Expected outcome:**
- Error `MalformedYAML` is returned.
