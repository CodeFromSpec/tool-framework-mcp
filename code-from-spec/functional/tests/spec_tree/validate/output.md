<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@DAhxZwsPSxR6jvDfIke2UdRt5BI -->

# SpecTreeValidate — Test Specification

## Interface

```
function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

record FormatError
  node: string
  rule: string
  detail: string
```

A node is a **leaf** if no other entry in the input list has a logical name
that starts with its logical name followed by `/`.
A node is **intermediate** if at least one such child entry exists.

---

## Happy path

### Test: Valid leaf node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter empty, node has name_section.heading = "ROOT", no agent, public section present
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT"], outputs = [{id: "out", path: "internal/out.go"}], node has name_section.heading = "ROOT/a", no agent

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns an empty list of FormatErrors.

---

### Test: Valid intermediate node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter empty, node has name_section.heading = "ROOT", no agent, public section present
- Entry 2: logical_name = "ROOT/a", frontmatter empty, node has name_section.heading = "ROOT/a", no agent, public section present (no agent section, no depends_on, no outputs, no input, no external)

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns an empty list of FormatErrors.

---

### Test: Leaf with no frontmatter fields

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter empty, node has name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter empty (no depends_on, no outputs, no input, no external), node has name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns an empty list of FormatErrors.

---

## Rule: name_heading

### Test: Heading matches logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "name_heading".

---

### Test: Heading does not match logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/wrong"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "name_heading"

---

## Rule: leaf_only_fields

### Test: Intermediate node with depends_on

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"
- (ROOT/a is intermediate because ROOT/a/b starts with "ROOT/a/")

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### Test: Intermediate node with outputs

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "x.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### Test: Intermediate node with input

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ARTIFACT/c(id)", node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### Test: Intermediate node with external

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "some/file.txt"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### Test: Intermediate node with multiple restricted fields

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"] and outputs = [{id: "x", path: "x.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns exactly two FormatErrors both with:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- (one for depends_on, one for outputs)

---

## Rule: leaf_only_agent

### Test: Intermediate node with agent section

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.agent present (non-empty), node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_agent"

---

### Test: Leaf node with agent section — no error

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.agent present, node.name_section.heading = "ROOT/a"
- (ROOT/a is a leaf — no entries start with "ROOT/a/")

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "leaf_only_agent".

---

## Rule: dependency_targets

### Test: depends_on targets non-existent ROOT node

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/missing"], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### Test: depends_on targets ancestor

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter has depends_on = ["ROOT"], node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a/b"
- rule = "dependency_targets"

---

### Test: depends_on targets descendant

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/a/b"], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", node.name_section.heading = "ROOT/a/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### Test: depends_on targets self

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/a"], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### Test: depends_on with valid ROOT qualifier

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a(interface)"], node.name_section.heading = "ROOT/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "dependency_targets".
(The qualifier "(interface)" is stripped; the resolved target "ROOT/a" exists and is neither ancestor, descendant, nor self of "ROOT/b".)

---

### Test: depends_on with valid ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "lib", path: "lib.go"}]
- Entry 3: logical_name = "ROOT/b", frontmatter has depends_on = ["ARTIFACT/a(lib)"], node.name_section.heading = "ROOT/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "dependency_targets".
(ARTIFACT/a(lib) resolves to node "ROOT/a" with output id "lib"; both exist.)

---

### Test: depends_on with non-existent ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ARTIFACT/missing(id)"], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### Test: Multiple invalid depends_on — one error per entry

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/missing", "ROOT/also_missing"], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns exactly two FormatErrors both with:
- node = "ROOT/a"
- rule = "dependency_targets"
- (one for "ROOT/missing", one for "ROOT/also_missing")

---

## Rule: input_target

### Test: Valid input reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "out", path: "a.go"}]
- Entry 3: logical_name = "ROOT/b", frontmatter has input = "ARTIFACT/a(out)", node.name_section.heading = "ROOT/b"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "input_target".

---

### Test: Input not starting with ARTIFACT/

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ROOT/something", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

### Test: Input references non-existent artifact

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ARTIFACT/missing(id)", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

## Rule: external_files

Fragment hashes use SHA-1 encoded as base64url (RFC 4648 §5, no padding) — always
27 characters. The hash input is each line in the declared range read with
`FileReadLine` (CRLF normalized to LF, terminator stripped), each with `\n` appended
— including the last line.

### Test: External file exists — no fragments

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "some/file.txt"}], node.name_section.heading = "ROOT/a"
- Create file "some/file.txt" on disk with any content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "external_files".

---

### Test: External file does not exist

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "nonexistent.txt"}], node.name_section.heading = "ROOT/a"
- Do not create "nonexistent.txt" on disk.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### Test: Fragment with valid hash

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "1-3", hash: <correct_hash>}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with 5 lines of known content, for example:
  - Line 1: "alpha"
  - Line 2: "beta"
  - Line 3: "gamma"
  - Line 4: "delta"
  - Line 5: "epsilon"
- Compute <correct_hash> as SHA-1 of the string "alpha\nbeta\ngamma\n" encoded as base64url without padding (27 characters).

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "external_files".

---

### Test: Fragment with invalid hash

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "1-3", hash: "wrong"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with at least 3 lines of known content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### Test: Fragment with invalid range format

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "abc", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### Test: Fragment with start greater than end

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "5-3", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### Test: Fragment with start less than 1

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "0-3", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### Test: Fragment out of range

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt", fragments: [{lines: "1-100", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with exactly 5 lines of content.

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates the fragment is out of range

---

## Rule: output_paths

### Test: Valid output path

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "internal/x.go"}], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "output_paths".

---

### Test: Output path with traversal

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "../../etc/passwd"}], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

### Test: Output path with backslash

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "internal\\x.go"}], node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

## Rule: duplicate_subsections

### Test: Unique subsection headings — no error

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains subsections with headings "Interface" and "Context", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

### Test: Duplicate subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains two subsections both named "Interface", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns exactly one FormatError where:
- node = "ROOT/a"
- rule = "duplicate_subsections"
- (the error corresponds to the second occurrence of "Interface")

---

### Test: Three identical subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains three subsections all named "Interface", node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns exactly two FormatErrors both with:
- node = "ROOT/a"
- rule = "duplicate_subsections"
- (second and third occurrences produce errors)

---

### Test: No public section — skip

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public absent, node.name_section.heading = "ROOT/a"

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

## Cross-cutting

### Test: Collects multiple errors from different rules

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", with all of the following:
  - node.name_section.heading = "ROOT/wrong" (triggers name_heading)
  - frontmatter has depends_on = ["ROOT/nonexistent"] (triggers dependency_targets)
  - node.public contains two subsections both named "Interface" (triggers duplicate_subsections)

**Action:** Call `SpecTreeValidate` with [Entry 1, Entry 2].

**Expected outcome:** Returns at least three FormatErrors with distinct rules:
- at least one with rule = "name_heading"
- at least one with rule = "dependency_targets"
- at least one with rule = "duplicate_subsections"

---

### Test: Empty input list

**Setup:** No entries.

**Action:** Call `SpecTreeValidate` with an empty list.

**Expected outcome:** Returns an empty list of FormatErrors.
