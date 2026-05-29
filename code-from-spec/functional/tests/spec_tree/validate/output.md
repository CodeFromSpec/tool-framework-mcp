<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@6kf8-p9niA2nFi8RLdRjTCq81Ro -->

# Test Specification: SpecTreeValidate

Function under test: `SpecTreeValidate`
Input type: `list of SpecTreeValidateInput`
Output type: `list of FormatError`

Each `SpecTreeValidateInput` has:
- `logical_name`: string
- `frontmatter`: Frontmatter (fields: `depends_on`, `outputs`, `input`, `external`)
- `node`: Node (fields: `name_section.heading`, `public`, `agent`)

Each `FormatError` has:
- `node`: string
- `rule`: string
- `detail`: string

A node is a **leaf** if no other entry in the input list has a logical name
that starts with that node's logical name followed by `/`.
A node is **intermediate** if at least one such entry exists.

---

## Happy Path

### TC-HP-1: Valid leaf node passes all checks

**Setup**
- Entry 1: `logical_name = "ROOT"`, frontmatter empty, heading = `"ROOT"`
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, depends_on = `["ROOT/b"]`
  (assume `ROOT/b` is also in the input), outputs = `[{id: "out", path: "internal/a.go"}]`

  For a self-contained variant, use:
- Entry 1: `logical_name = "ROOT"`, frontmatter empty, heading = `"ROOT"`
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty, no agent

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value is an empty list (no `FormatError` records).

---

### TC-HP-2: Valid intermediate node passes all checks

**Setup**
- Entry 1: `logical_name = "ROOT"`, frontmatter empty (no depends_on, outputs, input,
  external), heading = `"ROOT"`, no agent section
- Entry 2: `logical_name = "ROOT/a"`, frontmatter empty, heading = `"ROOT/a"`, no agent section

ROOT is intermediate (ROOT/a starts with "ROOT/"). ROOT/a is a leaf.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value is an empty list.

---

### TC-HP-3: Leaf with no frontmatter fields

**Setup**
- Entry 1: `logical_name = "ROOT"`, frontmatter empty, heading = `"ROOT"`
- Entry 2: `logical_name = "ROOT/a"`, frontmatter completely empty
  (depends_on absent, outputs absent, input absent, external absent),
  heading = `"ROOT/a"`, no agent section

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value is an empty list.

---

## Rule: name_heading

### TC-NH-1: Heading matches logical name — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "name_heading"`.

---

### TC-NH-2: Heading does not match logical name

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/wrong"`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains exactly one `FormatError` where:
- `node = "ROOT/a"`
- `rule = "name_heading"`

---

## Rule: leaf_only_fields

### TC-LOF-1: Intermediate node with depends_on

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, depends_on = `["ROOT/b"]`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty
- Entry 4: `logical_name = "ROOT/b"`, heading = `"ROOT/b"`, frontmatter empty

ROOT/a is intermediate (ROOT/a/b starts with "ROOT/a/").

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "leaf_only_fields"`

---

### TC-LOF-2: Intermediate node with outputs

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "x", path: "x.go"}]`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty

ROOT/a is intermediate.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "leaf_only_fields"`

---

### TC-LOF-3: Intermediate node with input

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  input = `"ARTIFACT/c(id)"`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty

ROOT/a is intermediate.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "leaf_only_fields"`

---

### TC-LOF-4: Intermediate node with external

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "some/file.txt"}]`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty

ROOT/a is intermediate.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "leaf_only_fields"`

---

### TC-LOF-5: Intermediate node with multiple restricted fields — one error per field

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ROOT/b"]`, outputs = `[{id: "x", path: "x.go"}]`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty
- Entry 4: `logical_name = "ROOT/b"`, heading = `"ROOT/b"`, frontmatter empty

ROOT/a is intermediate.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains exactly two `FormatError` records where both have:
- `node = "ROOT/a"`
- `rule = "leaf_only_fields"`

One error for `depends_on`, one error for `outputs`.

---

## Rule: leaf_only_agent

### TC-LOA-1: Intermediate node with agent section

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.agent` is present (non-absent)
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty

ROOT/a is intermediate.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "leaf_only_agent"`

---

### TC-LOA-2: Leaf node with agent section — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.agent` is present

ROOT/a is a leaf (no entry starts with "ROOT/a/").

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "leaf_only_agent"`.

---

## Rule: dependency_targets

### TC-DT-1: depends_on targets non-existent ROOT node

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ROOT/missing"]`

No entry with logical_name `"ROOT/missing"` exists.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "dependency_targets"`

---

### TC-DT-2: depends_on targets ancestor

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`,
  depends_on = `["ROOT"]`

`"ROOT"` is an ancestor of `"ROOT/a/b"` (ROOT/a/b starts with "ROOT/").

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a/b"`
- `rule = "dependency_targets"`

---

### TC-DT-3: depends_on targets descendant

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ROOT/a/b"]`
- Entry 3: `logical_name = "ROOT/a/b"`, heading = `"ROOT/a/b"`, frontmatter empty

`"ROOT/a/b"` is a descendant of `"ROOT/a"` (ROOT/a/b starts with "ROOT/a/").

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "dependency_targets"`

---

### TC-DT-4: depends_on targets self

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ROOT/a"]`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "dependency_targets"`

---

### TC-DT-5: depends_on with valid ROOT qualifier — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty
- Entry 3: `logical_name = "ROOT/b"`, heading = `"ROOT/b"`,
  depends_on = `["ROOT/a(interface)"]`

The qualifier `(interface)` is stripped; the resolved target is `"ROOT/a"`,
which exists and is not an ancestor, descendant, or self of `"ROOT/b"`.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "dependency_targets"`.

---

### TC-DT-6: depends_on with valid ARTIFACT reference — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "lib", path: "lib.go"}]`
- Entry 3: `logical_name = "ROOT/b"`, heading = `"ROOT/b"`,
  depends_on = `["ARTIFACT/a(lib)"]`

The reference `"ARTIFACT/a(lib)"` means: node `"ROOT/a"` must have an output
with id `"lib"`. It does.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "dependency_targets"`.

---

### TC-DT-7: depends_on with non-existent ARTIFACT reference

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ARTIFACT/missing(id)"]`

No entry with logical_name `"ROOT/missing"` exists.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "dependency_targets"`

---

### TC-DT-8: Multiple invalid depends_on — one error per entry

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  depends_on = `["ROOT/missing", "ROOT/also_missing"]`

Neither `"ROOT/missing"` nor `"ROOT/also_missing"` exist in the input list.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains exactly two `FormatError` records where both have:
- `node = "ROOT/a"`
- `rule = "dependency_targets"`

One error per invalid depends_on entry.

---

## Rule: input_target

### TC-IT-1: Valid input reference — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "out", path: "a.go"}]`
- Entry 3: `logical_name = "ROOT/b"`, heading = `"ROOT/b"`,
  input = `"ARTIFACT/a(out)"`

Node `"ROOT/a"` exists and has output with id `"out"`.

**Actions**
Call `SpecTreeValidate` with all entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "input_target"`.

---

### TC-IT-2: Input not starting with "ARTIFACT/"

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  input = `"ROOT/something"`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "input_target"`

---

### TC-IT-3: Input references non-existent artifact

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  input = `"ARTIFACT/missing(id)"`

No entry with logical_name `"ROOT/missing"` exists.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "input_target"`

---

## Rule: external_files

Hash algorithm: SHA-1 of the selected lines joined with `\n` (LF), where each
line is read with `FileReadLine` (CRLF normalized to LF, line terminators
stripped). The digest is encoded as base64url (RFC 4648 §5, no padding) —
always 27 characters.

### TC-EF-1: External file exists — no fragments

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "some/file.txt"}]`
- Create the file `"some/file.txt"` on disk with any content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "external_files"`.

---

### TC-EF-2: External file does not exist

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "nonexistent.txt"}]`
- Do not create the file `"nonexistent.txt"` on disk.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`

---

### TC-EF-3: Fragment with valid hash

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "1-3", hash: <correct-hash>}]}]`
- Create the file `"f.txt"` on disk with at least 5 lines of known content,
  for example:
  ```
  line one
  line two
  line three
  line four
  line five
  ```
- Compute `<correct-hash>` by:
  1. Reading lines 1–3 with `FileReadLine` (stripping terminators, normalizing CRLF).
  2. Joining them with `\n`: `"line one\nline two\nline three"`
  3. Computing SHA-1 of that string.
  4. Encoding the digest as base64url (RFC 4648 §5, no padding, 27 characters).

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "external_files"`.

---

### TC-EF-4: Fragment with invalid hash

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "1-3", hash: "wrong"}]}]`
- Create `"f.txt"` on disk with at least 3 lines of known content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`

---

### TC-EF-5: Fragment with invalid range format

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "abc", hash: "x"}]}]`
- Create `"f.txt"` on disk with any content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`

---

### TC-EF-6: Fragment with start > end

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "5-3", hash: "x"}]}]`
- Create `"f.txt"` on disk with any content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`

---

### TC-EF-7: Fragment with start < 1

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "0-3", hash: "x"}]}]`
- Create `"f.txt"` on disk with any content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`

---

### TC-EF-8: Fragment out of range

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  external = `[{path: "f.txt", fragments: [{lines: "1-100", hash: "x"}]}]`
- Create `"f.txt"` on disk with exactly 5 lines of content.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "external_files"`
- `detail` indicates the fragment is out of range.

---

## Rule: output_paths

### TC-OP-1: Valid output path — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "x", path: "internal/x.go"}]`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "output_paths"`.

---

### TC-OP-2: Output path with traversal

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "x", path: "../../etc/passwd"}]`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "output_paths"`

---

### TC-OP-3: Output path with backslash

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`,
  outputs = `[{id: "x", path: "internal\\x.go"}]`

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains a `FormatError` where:
- `node = "ROOT/a"`
- `rule = "output_paths"`

---

## Rule: duplicate_subsections

### TC-DS-1: Unique subsection headings — no error

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.public` contains subsections with headings `"Interface"` and `"Context"`
  (all distinct).

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "duplicate_subsections"`.

---

### TC-DS-2: Duplicate subsection headings

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.public` contains two subsections both named `"Interface"`.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains exactly one `FormatError` where:
- `node = "ROOT/a"`
- `rule = "duplicate_subsections"`

The error corresponds to the second occurrence of `"Interface"`.

---

### TC-DS-3: Three identical subsection headings

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.public` contains three subsections all named `"Interface"`.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains exactly two `FormatError` records where both have:
- `node = "ROOT/a"`
- `rule = "duplicate_subsections"`

The two errors correspond to the second and third occurrences of `"Interface"`.

---

### TC-DS-4: No public section — skip

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, frontmatter empty,
  `node.public` is absent.

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains no `FormatError` with `rule = "duplicate_subsections"`.

---

## Cross-Cutting

### TC-CC-1: Collects multiple errors from different rules

**Setup**
- Entry 1: `logical_name = "ROOT"`, heading = `"ROOT"`, frontmatter empty
- Entry 2: `logical_name = "ROOT/a"`, heading = `"ROOT/a"`, all of:
  - `node.name_section.heading = "ROOT/wrong"` (triggers `name_heading`)
  - `depends_on = ["ROOT/missing"]` (triggers `dependency_targets`)
  - `node.public` contains two subsections both named `"Interface"`
    (triggers `duplicate_subsections`)

**Actions**
Call `SpecTreeValidate` with both entries.

**Expected outcome**
Return value contains at least three `FormatError` records. At least one has
`rule = "name_heading"`, at least one has `rule = "dependency_targets"`, and
at least one has `rule = "duplicate_subsections"`. All have `node = "ROOT/a"`.

---

### TC-CC-2: Empty input list

**Setup**
Input is an empty list (no `SpecTreeValidateInput` records).

**Actions**
Call `SpecTreeValidate` with an empty list.

**Expected outcome**
Return value is an empty list (no `FormatError` records).
