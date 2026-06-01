<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@aNIQvNF-HLwdbeJPvfv1OEiHM9o -->

# Test Specification: SpecTreeValidate

This document describes test cases for the `SpecTreeValidate` function. Each
test lists its input (a list of `SpecTreeValidateInput` records), the function
call performed, and the expected outcome.

A node is considered to have children if any other entry in the input list has
a logical name that starts with that node's logical name followed by `/`. A
node is a leaf if no entry starts with its logical name followed by `/`.

Fragment hashes use SHA-1 encoded as base64url (RFC 4648 §5, no padding) —
always 27 characters. The input to SHA-1 is each line in the declared range,
read with CRLF normalized to LF and terminators stripped, then each line
re-appended with `\n` (LF), including the last line.

---

## Happy Path

### Test: Valid leaf node passes all checks

**Setup:**
- Entry `ROOT`: intermediate node (has children `ROOT/a` and `ROOT/b`).
  `node.name_section.heading` = `"ROOT"`. Empty frontmatter.
- Entry `ROOT/a`: leaf node. `node.name_section.heading` = `"ROOT/a"`.
  `depends_on` = `["ROOT/b"]`. `outputs` = `[{id: "out", path: "internal/out.go"}]`.
- Entry `ROOT/b`: leaf node. `node.name_section.heading` = `"ROOT/b"`.
  Empty frontmatter.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns an empty list of `FormatError`. No errors.

---

### Test: Valid intermediate node passes all checks

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
  `node.name_section.heading` = `"ROOT"`. `node.public` present with empty
  content. No frontmatter fields. No agent section.
- Entry `ROOT/a`: leaf node. `node.name_section.heading` = `"ROOT/a"`.
  Empty frontmatter.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns an empty list of `FormatError`. No errors.

---

### Test: Leaf with no frontmatter fields

**Setup:**
- Entry `ROOT`: node with `node.name_section.heading` = `"ROOT"`.
- Entry `ROOT/a`: leaf node. `node.name_section.heading` = `"ROOT/a"`.
  Empty frontmatter (no `depends_on`, no `outputs`, no `input`, no `external`).

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns an empty list of `FormatError`. No errors.

---

## Rule: name_heading

### Test: Heading matches logical name

**Setup:**
- Entry `ROOT`: `node.name_section.heading` = `"ROOT"`.
- Entry `ROOT/a`: `node.name_section.heading` = `"ROOT/a"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` = `"name_heading"`.

---

### Test: Heading does not match logical name

**Setup:**
- Entry `ROOT`: `node.name_section.heading` = `"ROOT"`.
- Entry `ROOT/a`: `node.name_section.heading` = `"ROOT/wrong"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"name_heading"`

---

## Rule: leaf_only_fields

### Test: Intermediate node with depends_on

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `depends_on` = `["ROOT/b"]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"leaf_only_fields"`

---

### Test: Intermediate node with outputs

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `outputs` = `[{id: "x", path: "x.go"}]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"leaf_only_fields"`

---

### Test: Intermediate node with input

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `input` = `"ARTIFACT/c(id)"`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"leaf_only_fields"`

---

### Test: Intermediate node with external

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `external` = `[{path: "some/file.txt"}]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"leaf_only_fields"`

---

### Test: Intermediate node with multiple restricted fields

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `depends_on` = `["ROOT/b"]`. `outputs` = `[{id: "x", path: "x.go"}]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns exactly two `FormatError` records where both
have `node` = `"ROOT/a"` and `rule` = `"leaf_only_fields"` — one per
restricted field (`depends_on` and `outputs`).

---

## Rule: leaf_only_agent

### Test: Intermediate node with agent section

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `node.agent` present with `content` = `["Agent instructions."]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"leaf_only_agent"`

---

### Test: Leaf node with agent section — no error

**Setup:**
- Entry `ROOT`: `node.name_section.heading` = `"ROOT"`.
- Entry `ROOT/a`: leaf node. `node.name_section.heading` = `"ROOT/a"`.
  `node.agent` present with `content` = `["Agent instructions."]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` = `"leaf_only_agent"`.

---

## Rule: dependency_targets

### Test: depends_on targets non-existent ROOT node

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `depends_on` = `["ROOT/missing"]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"dependency_targets"`

---

### Test: depends_on targets ancestor

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
- Entry `ROOT/a/b`: leaf node. `depends_on` = `["ROOT"]`.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a/b"`
- `rule` = `"dependency_targets"`

---

### Test: depends_on targets descendant

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: intermediate node (has child `ROOT/a/b`).
  `depends_on` = `["ROOT/a/b"]`.
- Entry `ROOT/a/b`: leaf node.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"dependency_targets"`

---

### Test: depends_on targets self

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `depends_on` = `["ROOT/a"]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"dependency_targets"`

---

### Test: depends_on with valid ROOT qualifier

**Setup:**
- Entry `ROOT`: intermediate node (has children `ROOT/a` and `ROOT/b`).
- Entry `ROOT/a`: leaf node. Empty frontmatter.
- Entry `ROOT/b`: leaf node. `depends_on` = `["ROOT/a(interface)"]`.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns no `FormatError` with `rule` =
`"dependency_targets"`. The qualifier `(interface)` is stripped before
resolving; `ROOT/a` exists and is neither ancestor, descendant, nor self of
`ROOT/b`.

---

### Test: depends_on with valid ARTIFACT reference

**Setup:**
- Entry `ROOT`: intermediate node (has children `ROOT/a` and `ROOT/b`).
- Entry `ROOT/a`: leaf node. `outputs` = `[{id: "lib", path: "lib.go"}]`.
- Entry `ROOT/b`: leaf node. `depends_on` = `["ARTIFACT/a(lib)"]`.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns no `FormatError` with `rule` =
`"dependency_targets"`. `ARTIFACT/a(lib)` resolves to node `ROOT/a` with
output id `"lib"`, which exists.

---

### Test: depends_on with non-existent ARTIFACT reference

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `depends_on` = `["ARTIFACT/missing(id)"]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"dependency_targets"`

---

### Test: Multiple invalid depends_on — one error per entry

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `depends_on` = `["ROOT/missing", "ROOT/also_missing"]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns exactly two `FormatError` records, both with
`node` = `"ROOT/a"` and `rule` = `"dependency_targets"` — one per invalid
dependency entry.

---

## Rule: input_target

### Test: Valid input reference

**Setup:**
- Entry `ROOT`: intermediate node (has children `ROOT/a` and `ROOT/b`).
- Entry `ROOT/a`: leaf node. `outputs` = `[{id: "out", path: "a.go"}]`.
- Entry `ROOT/b`: leaf node. `input` = `"ARTIFACT/a(out)"`.

**Action:** Call `SpecTreeValidate` with all three entries.

**Expected outcome:** Returns no `FormatError` with `rule` = `"input_target"`.

---

### Test: Input not starting with ARTIFACT/

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `input` = `"ROOT/something"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"input_target"`

---

### Test: Input references non-existent artifact

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `input` = `"ARTIFACT/missing(id)"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"input_target"`

---

## Rule: external_files

### Test: External file exists — no fragments

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `external` = `[{path: "some/file.txt"}]`.
- File on disk: create `"some/file.txt"` with content `"hello\n"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` = `"external_files"`.

---

### Test: External file does not exist

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `external` = `[{path: "nonexistent.txt"}]`.
- File on disk: do not create `"nonexistent.txt"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`

---

### Test: Fragment with valid hash

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "1-3", hash: <correct_hash>}]}]`.
- File on disk: create `"f.txt"` with exactly 5 lines:
  ```
  alpha
  bravo
  charlie
  delta
  echo
  ```
- Compute `<correct_hash>`: apply SHA-1 to the bytes produced by taking lines
  1, 2, and 3 (`"alpha"`, `"bravo"`, `"charlie"`), each appended with `\n`
  (LF), concatenated. Encode the SHA-1 digest as base64url without padding
  (RFC 4648 §5) to produce a 27-character string.

**Action:** Call `SpecTreeValidate` with both entries using the computed hash.

**Expected outcome:** Returns no `FormatError` with `rule` = `"external_files"`.

---

### Test: Fragment with invalid hash

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "1-3", hash: "wrong_______________________"}]}]`.
- File on disk: create `"f.txt"` with exactly 5 lines:
  ```
  alpha
  bravo
  charlie
  delta
  echo
  ```

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`

---

### Test: Fragment with invalid range format

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "abc", hash: "x"}]}]`.
- File on disk: create `"f.txt"` with content `"hello\n"`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`

---

### Test: Fragment with start > end

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "5-3", hash: "x"}]}]`.
- File on disk: create `"f.txt"` with exactly 5 lines:
  ```
  alpha
  bravo
  charlie
  delta
  echo
  ```

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`

---

### Test: Fragment with start < 1

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "0-3", hash: "x"}]}]`.
- File on disk: create `"f.txt"` with exactly 5 lines:
  ```
  alpha
  bravo
  charlie
  delta
  echo
  ```

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`

---

### Test: Fragment out of range

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `external` = `[{path: "f.txt", fragments: [{lines: "1-100", hash: "x"}]}]`.
- File on disk: create `"f.txt"` with exactly 5 lines:
  ```
  alpha
  bravo
  charlie
  delta
  echo
  ```

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"external_files"`
- `detail` indicates the fragment range exceeds the file's line count.

---

## Rule: output_paths

### Test: Valid output path

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `outputs` = `[{id: "x", path: "internal/x.go"}]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` = `"output_paths"`.

---

### Test: Output path with traversal

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `outputs` = `[{id: "x", path: "../../etc/passwd"}]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"output_paths"`

---

### Test: Output path with backslash

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node.
  `outputs` = `[{id: "x", path: "internal\\x.go"}]`.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"output_paths"`

---

## Rule: duplicate_subsections

### Test: Unique subsection headings — no error

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `node.public` present with subsections:
  - `{heading: "interface", raw_heading: "## Interface", content: ["Types."]}`
  - `{heading: "context", raw_heading: "## Context", content: ["Background."]}`

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` =
`"duplicate_subsections"`.

---

### Test: Duplicate subsection headings

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `node.public` present with subsections:
  - `{heading: "interface", raw_heading: "## Interface", content: ["First."]}`
  - `{heading: "interface", raw_heading: "## Interface", content: ["Second."]}`

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns exactly one `FormatError` where:
- `node` = `"ROOT/a"`
- `rule` = `"duplicate_subsections"`
The error corresponds to the second occurrence of the heading `"interface"`.

---

### Test: Three identical subsection headings

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `node.public` present with subsections:
  - `{heading: "interface", raw_heading: "## Interface", content: ["First."]}`
  - `{heading: "interface", raw_heading: "## Interface", content: ["Second."]}`
  - `{heading: "interface", raw_heading: "## Interface", content: ["Third."]}`

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns exactly two `FormatError` records, both with:
- `node` = `"ROOT/a"`
- `rule` = `"duplicate_subsections"`
The errors correspond to the second and third occurrences of heading
`"interface"`.

---

### Test: No public section — skip

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node. `node.public` absent.

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns no `FormatError` with `rule` =
`"duplicate_subsections"`.

---

## Cross-Cutting

### Test: Collects multiple errors from different rules

**Setup:**
- Entry `ROOT`: intermediate node (has child `ROOT/a`).
- Entry `ROOT/a`: leaf node with the following issues:
  - `node.name_section.heading` = `"ROOT/wrong"` (violates `name_heading`)
  - `depends_on` = `["ROOT/missing"]` (violates `dependency_targets`)
  - `node.public` present with subsections:
    - `{heading: "interface", raw_heading: "## Interface", content: ["First."]}`
    - `{heading: "interface", raw_heading: "## Interface", content: ["Second."]}`
    (violates `duplicate_subsections`)

**Action:** Call `SpecTreeValidate` with both entries.

**Expected outcome:** Returns at least three `FormatError` records:
- At least one with `node` = `"ROOT/a"` and `rule` = `"name_heading"`
- At least one with `node` = `"ROOT/a"` and `rule` = `"dependency_targets"`
- At least one with `node` = `"ROOT/a"` and `rule` = `"duplicate_subsections"`

---

### Test: Empty input list

**Setup:** No entries — pass an empty list to `SpecTreeValidate`.

**Action:** Call `SpecTreeValidate` with an empty list.

**Expected outcome:** Returns an empty list of `FormatError`. No errors.
