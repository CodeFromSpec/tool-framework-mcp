<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@1bQW4fKwtmpmY1d0pS44FQ5EAUM -->

# Test Specification: SpecTreeValidate

## Interface

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

record FormatError
  node: string
  rule: string
  detail: string

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError
```

---

## Happy Path

### TC-HP-1: Valid leaf node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node has name_section.heading = "ROOT", no agent section, no restricted fields
- Entry 2: logical_name = "ROOT/a", frontmatter with valid depends_on = ["ROOT/b"] (where ROOT/b exists), valid outputs = [{id: "x", path: "x.go"}], node has name_section.heading = "ROOT/a", no agent section

  (Add Entry 3: logical_name = "ROOT/b" to satisfy depends_on target)

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns empty list (no FormatErrors).

---

### TC-HP-2: Valid intermediate node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node has name_section.heading = "ROOT", only name section and public section, no frontmatter fields, no agent section
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node has name_section.heading = "ROOT/a", only name section and public section, no agent section

  (ROOT is intermediate because ROOT/a starts with "ROOT/")

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns empty list (no FormatErrors).

---

### TC-HP-3: Leaf with no frontmatter fields

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node has name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty (no depends_on, no outputs, no input, no external), node has name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns empty list (no FormatErrors).

---

## Rule: name_heading

### TC-NH-1: Heading matches logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "name_heading".

---

### TC-NH-2: Heading does not match logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/wrong"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "name_heading"

---

## Rule: leaf_only_fields

### TC-LOF-1: Intermediate node with depends_on

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/b"], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"
- Entry 4: logical_name = "ROOT/b", frontmatter = empty, node.name_section.heading = "ROOT/b"

  (ROOT/a is intermediate because ROOT/a/b starts with "ROOT/a/")

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### TC-LOF-2: Intermediate node with outputs

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "x", path: "x.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### TC-LOF-3: Intermediate node with input

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with input = "ARTIFACT/c(id)", node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### TC-LOF-4: Intermediate node with external

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "some/file.txt"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

### TC-LOF-5: Intermediate node with multiple restricted fields

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/b"] and outputs = [{id: "x", path: "x.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"
- Entry 4: logical_name = "ROOT/b", frontmatter = empty, node.name_section.heading = "ROOT/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "leaf_only_fields"
  (one for depends_on, one for outputs)

---

## Rule: leaf_only_agent

### TC-LOA-1: Intermediate node with agent section

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.agent is present (non-empty)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_agent"

---

### TC-LOA-2: Leaf node with agent section — no error

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.agent is present (non-empty)

  (ROOT/a is a leaf — no entry starts with "ROOT/a/")

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "leaf_only_agent".

---

## Rule: dependency_targets

### TC-DT-1: depends_on targets non-existent ROOT node

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/missing"], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### TC-DT-2: depends_on targets ancestor

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter with depends_on = ["ROOT"], node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a/b"
- rule = "dependency_targets"

---

### TC-DT-3: depends_on targets descendant

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/a/b"], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter = empty, node.name_section.heading = "ROOT/a/b"

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### TC-DT-4: depends_on targets self

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/a"], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### TC-DT-5: depends_on with valid ROOT qualifier

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter with depends_on = ["ROOT/a(interface)"], node.name_section.heading = "ROOT/b"

  (qualifier "(interface)" is stripped before resolving — target becomes "ROOT/a" which exists and is not ancestor/descendant/self of ROOT/b)

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** No FormatError with rule = "dependency_targets".

---

### TC-DT-6: depends_on with valid ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "lib", path: "lib.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter with depends_on = ["ARTIFACT/a(lib)"], node.name_section.heading = "ROOT/b"

  (ARTIFACT/a(lib) resolves: node = "ROOT/a", output id = "lib"; ROOT/a exists and has output with id "lib")

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** No FormatError with rule = "dependency_targets".

---

### TC-DT-7: depends_on with non-existent ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ARTIFACT/missing(id)"], node.name_section.heading = "ROOT/a"

  ("ROOT/missing" does not exist in the entry list)

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

### TC-DT-8: Multiple invalid depends_on — one error per entry

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/missing", "ROOT/also_missing"], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "dependency_targets"
  (one for "ROOT/missing", one for "ROOT/also_missing")

---

## Rule: input_target

### TC-IT-1: Valid input reference

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "out", path: "a.go"}], node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter with input = "ARTIFACT/a(out)", node.name_section.heading = "ROOT/b"

  (ARTIFACT/a(out) resolves: node = "ROOT/a", output id = "out"; ROOT/a exists and has output with id "out")

**Action:** Call SpecTreeValidate with all entries.

**Expected outcome:** No FormatError with rule = "input_target".

---

### TC-IT-2: Input not starting with "ARTIFACT/"

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with input = "ROOT/something", node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

### TC-IT-3: Input references non-existent artifact

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with input = "ARTIFACT/missing(id)", node.name_section.heading = "ROOT/a"

  ("ROOT/missing" does not exist in the entry list)

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

## Rule: external_files

**Hash algorithm note:** Fragment hashes use SHA-1 encoded as base64url (RFC 4648 §5, no padding) — always 27 characters. The SHA-1 input is each line in the declared range read with FileReadLine (CRLF normalized to LF, terminator stripped), with "\n" appended to each line including the last.

### TC-EF-1: External file exists — no fragments

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "some/file.txt"}], node.name_section.heading = "ROOT/a"
- Create file "some/file.txt" on disk with any content.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "external_files".

---

### TC-EF-2: External file does not exist

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "nonexistent.txt"}], node.name_section.heading = "ROOT/a"
- Do not create "nonexistent.txt" on disk.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### TC-EF-3: Fragment with valid hash

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "1-3", hash: <correct-hash>}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with 5 lines of known content, e.g.:
  - Line 1: "alpha"
  - Line 2: "beta"
  - Line 3: "gamma"
  - Line 4: "delta"
  - Line 5: "epsilon"
- Compute <correct-hash>: SHA-1 of "alpha\nbeta\ngamma\n", then encode as base64url with no padding (27 characters).

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "external_files".

---

### TC-EF-4: Fragment with invalid hash

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "1-3", hash: "wrong"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with at least 3 lines of known content.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### TC-EF-5: Fragment with invalid range format

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "abc", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### TC-EF-6: Fragment with start > end

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "5-3", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### TC-EF-7: Fragment with start < 1

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "0-3", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with any content.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

### TC-EF-8: Fragment out of range

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with external = [{path: "f.txt", fragments: [{lines: "1-100", hash: "x"}]}], node.name_section.heading = "ROOT/a"
- Create file "f.txt" on disk with exactly 5 lines.

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates fragment out of range

---

## Rule: output_paths

### TC-OP-1: Valid output path

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "x", path: "internal/x.go"}], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "output_paths".

---

### TC-OP-2: Output path with traversal

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "x", path: "../../etc/passwd"}], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

### TC-OP-3: Output path with backslash

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with outputs = [{id: "x", path: "internal\\x.go"}], node.name_section.heading = "ROOT/a"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

## Rule: duplicate_subsections

### TC-DS-1: Unique subsection headings — no error

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.public contains two subsections with headings "Interface" and "Context"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

### TC-DS-2: Duplicate subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.public contains two subsections both named "Interface"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly one FormatError where:
- node = "ROOT/a"
- rule = "duplicate_subsections"
  (the second occurrence of "Interface")

---

### TC-DS-3: Three identical subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.public contains three subsections all named "Interface"

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "duplicate_subsections"
  (second and third occurrences of "Interface")

---

### TC-DS-4: No public section — skip

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = empty, node.name_section.heading = "ROOT/a", node.public is absent

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

## Cross-Cutting

### TC-CC-1: Collects multiple errors from different rules

**Setup:**
- Entry 1: logical_name = "ROOT", frontmatter = empty, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter with depends_on = ["ROOT/missing"], node.name_section.heading = "ROOT/wrong", node.public contains two subsections both named "Interface"

  (three violations: name_heading mismatch, invalid depends_on, duplicate subsections)

**Action:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least three FormatErrors covering at least three distinct rule values: "name_heading", "dependency_targets", "duplicate_subsections".

---

### TC-CC-2: Empty input list

**Setup:** No entries.

**Action:** Call SpecTreeValidate with an empty list.

**Expected outcome:** Returns empty list (no FormatErrors).
