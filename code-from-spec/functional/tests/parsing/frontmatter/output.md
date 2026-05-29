<!-- code-from-spec: ROOT/functional/tests/parsing/frontmatter@Y0VFPHy4f98dzStomPukGYRwSa8 -->

# Frontmatter Parse — Test Specification

## Data Types

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

---

## Happy Path

---

### Test: Parses complete frontmatter (all fields)

**Setup**
Create a file whose content is:

```
---
depends_on:
  - "dep-one"
  - "dep-two"
external:
  - path: "some/external/file.md"
    fragments:
      - description: "Fragment description"
        lines: "10-20"
        hash: "abc123"
input: "some/input/file.md"
outputs:
  - id: "out-one"
    path: "path/to/out-one.go"
  - id: "out-two"
    path: "path/to/out-two.go"
---
body content here
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` contains `"dep-one"` and `"dep-two"`.
`external` has one entry with `path` equal to `"some/external/file.md"` and one fragment with `description` equal to `"Fragment description"`, `lines` equal to `"10-20"`, and `hash` equal to `"abc123"`.
`input` equals `"some/input/file.md"`.
`outputs` has two entries: first with `id` `"out-one"` and `path` `"path/to/out-one.go"`, second with `id` `"out-two"` and `path` `"path/to/out-two.go"`.

---

### Test: Parses frontmatter with only outputs

**Setup**
Create a file whose content is:

```
---
outputs:
  - id: "result"
    path: "gen/result.go"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` is empty.
`external` is empty.
`input` is empty.
`outputs` has one entry with `id` `"result"` and `path` `"gen/result.go"`.

---

### Test: Parses frontmatter with only depends_on

**Setup**
Create a file whose content is:

```
---
depends_on:
  - "alpha"
  - "beta"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` contains `"alpha"` and `"beta"`.
`external` is empty.
`input` is empty.
`outputs` is empty.

---

### Test: Parses frontmatter with external and fragments

**Setup**
Create a file whose content is:

```
---
external:
  - path: "first/path.md"
    fragments:
      - description: "First fragment"
        lines: "1-5"
        hash: "hash-one"
      - description: "Second fragment"
        lines: "7-9"
        hash: "hash-two"
  - path: "second/path.md"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`external` has two entries.
First entry has `path` `"first/path.md"` and two fragments:
  - fragment one: `description` `"First fragment"`, `lines` `"1-5"`, `hash` `"hash-one"`.
  - fragment two: `description` `"Second fragment"`, `lines` `"7-9"`, `hash` `"hash-two"`.
Second entry has `path` `"second/path.md"` and no fragments.

---

### Test: Parses frontmatter with input field

**Setup**
Create a file whose content is:

```
---
input: "data/source.txt"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`input` equals `"data/source.txt"`.
`depends_on` is empty.
`external` is empty.
`outputs` is empty.

---

### Test: Fragment without description

**Setup**
Create a file whose content is:

```
---
external:
  - path: "some/file.md"
    fragments:
      - lines: "3-8"
        hash: "no-desc-hash"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`external` has one entry with one fragment.
The fragment has `description` absent, `lines` equal to `"3-8"`, and `hash` equal to `"no-desc-hash"`.

---

### Test: Ignores unknown frontmatter fields

**Setup**
Create a file whose content is:

```
---
depends_on:
  - "known-dep"
custom_field: "ignored value"
another_unknown: 42
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` contains `"known-dep"`.
`external` is empty.
`input` is empty.
`outputs` is empty.
Unknown fields `custom_field` and `another_unknown` are silently ignored.

---

### Test: File with no frontmatter returns empty result

**Setup**
Create a file whose content contains no `---` delimiter — body text only, for example:

```
This is just body content.
No frontmatter here.
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` is empty.
`external` is empty.
`input` is empty.
`outputs` is empty.

---

## Edge Cases

---

### Test: Empty frontmatter

**Setup**
Create a file whose content is:

```
---
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` is empty.
`external` is empty.
`input` is empty.
`outputs` is empty.

---

### Test: File with only frontmatter, nothing after

**Setup**
Create a file whose content is:

```
---
depends_on:
  - "lonely-dep"
---
```
(No body content after the closing `---`.)

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
`depends_on` contains `"lonely-dep"`.
All other fields are empty.

---

### Test: Delimiter with trailing whitespace is not recognized

**Setup**
Create a file whose first line is `---   ` (three dashes followed by spaces), followed by body content. No valid `---` delimiter exists.

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
No error.
Result has all fields empty — the line with trailing spaces is not recognized as a delimiter.

---

## Failure Cases

---

### Test: File does not exist

**Setup**
No file is created. Use a `PathCfs` that points to a non-existent file.

**Action**
Call `FrontmatterParse` with that path.

**Expected outcome**
Error `"file unreadable"`.

---

### Test: Propagates path errors

**Setup**
No file is created. Use a `PathCfs` constructed from a path that attempts directory traversal, for example `"../../outside"`.

**Action**
Call `FrontmatterParse` with that path.

**Expected outcome**
Error `"directory traversal"` propagated from `FileOpen`.

---

### Test: Malformed YAML

**Setup**
Create a file whose content is:

```
---
depends_on: [unclosed bracket
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
Error `"malformed YAML"`.

---

### Test: Unclosed frontmatter block

**Setup**
Create a file that begins with `---` but has no second `---` line:

```
---
depends_on:
  - "something"
body content with no closing delimiter
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
Error `"malformed YAML"`.

---

### Test: Missing required field in external entry

**Setup**
Create a file with an `external` entry that has no `path` field:

```
---
external:
  - fragments:
      - lines: "1-2"
        hash: "abc"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
Error `"malformed YAML"`.

---

### Test: Missing required field in fragment

**Setup**
Create a file with a fragment that has `lines` but no `hash`:

```
---
external:
  - path: "some/file.md"
    fragments:
      - lines: "1-5"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
Error `"malformed YAML"`.

---

### Test: Missing required field in output entry

**Setup**
Create a file with an `outputs` entry that has `id` but no `path`:

```
---
outputs:
  - id: "out-only-id"
---
```

**Action**
Call `FrontmatterParse` with the path to that file.

**Expected outcome**
Error `"malformed YAML"`.
