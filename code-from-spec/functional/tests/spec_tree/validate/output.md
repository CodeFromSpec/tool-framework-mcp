<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@1bQW4fKwtmpmY1d0pS44FQ5EAUM -->

# Test Specification: SpecTreeValidate

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

A node is considered a leaf if no other entry in the input list has a logical
name that starts with that node's logical name followed by "/". A node is
intermediate if at least one such entry exists.

Fragment hashes use SHA-1 encoded as base64url (RFC 4648 §5, no padding),
always 27 characters. The SHA-1 input is each line in the declared range read
with FileReadLine (CRLF normalized to LF, terminators stripped), with "\n"
appended — including the last line.

---

## Test Cases

### Happy Path

---

#### TC-HP-1: Valid leaf node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", node has name_section.heading = "ROOT",
  no frontmatter fields, no agent section.
- Entry 2: logical_name = "ROOT/a", node has name_section.heading = "ROOT/a",
  frontmatter has depends_on = ["ROOT/b"], outputs = [{id: "out", path: "out.go"}],
  no agent section. (ROOT/b included below.)
- Entry 3: logical_name = "ROOT/b", node has name_section.heading = "ROOT/b",
  no frontmatter fields.

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns an empty list of FormatErrors.

---

#### TC-HP-2: Valid intermediate node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", node has name_section.heading = "ROOT",
  no frontmatter fields, no agent section, node.public present with unique
  subsections.
- Entry 2: logical_name = "ROOT/a", node has name_section.heading = "ROOT/a",
  no frontmatter fields, no agent section.

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns an empty list of FormatErrors.

---

#### TC-HP-3: Leaf with no frontmatter fields

**Setup:**
- Entry 1: logical_name = "ROOT", node has name_section.heading = "ROOT".
- Entry 2: logical_name = "ROOT/a", node has name_section.heading = "ROOT/a",
  frontmatter is empty (no depends_on, no outputs, no input, no external).

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns an empty list of FormatErrors.

---

### Rule: name_heading

---

#### TC-NH-1: Heading matches logical name — no error

**Setup:**
- Entry 1: logical_name = "ROOT", node has name_section.heading = "ROOT".
- Entry 2: logical_name = "ROOT/a", node has name_section.heading = "ROOT/a".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "name_heading".

---

#### TC-NH-2: Heading does not match logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node has name_section.heading = "ROOT".
- Entry 2: logical_name = "ROOT/a", node has name_section.heading = "ROOT/wrong".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "name_heading"

---

### Rule: leaf_only_fields

---

#### TC-LOF-1: Intermediate node with depends_on

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"].
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".
- Entry 4: logical_name = "ROOT/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3, Entry 4].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### TC-LOF-2: Intermediate node with outputs

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x", path: "x.go"}].
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### TC-LOF-3: Intermediate node with input

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ARTIFACT/c(id)".
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".
- Entry 4: logical_name = "ROOT/c", frontmatter has outputs = [{id: "id", path: "c.go"}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3, Entry 4].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### TC-LOF-4: Intermediate node with external

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "some/file.txt"}].
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### TC-LOF-5: Intermediate node with multiple restricted fields — one error per field

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/b"]
  AND outputs = [{id: "x", path: "x.go"}].
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".
- Entry 4: logical_name = "ROOT/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3, Entry 4].

**Expected outcome:** Returns at least two FormatErrors, each with:
- node = "ROOT/a"
- rule = "leaf_only_fields"
One error for the depends_on field and one for the outputs field.

---

### Rule: leaf_only_agent

---

#### TC-LOA-1: Intermediate node with agent section

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.agent is present.
  (ROOT/a is intermediate because ROOT/a/b exists.)
- Entry 3: logical_name = "ROOT/a/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_agent"

---

#### TC-LOA-2: Leaf node with agent section — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.agent is present.
  (ROOT/a is a leaf — no entry starts with "ROOT/a/".)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "leaf_only_agent".

---

### Rule: dependency_targets

---

#### TC-DT-1: depends_on targets non-existent ROOT node

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/missing"].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### TC-DT-2: depends_on targets ancestor

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a".
- Entry 3: logical_name = "ROOT/a/b", frontmatter has depends_on = ["ROOT"].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a/b"
- rule = "dependency_targets"

---

#### TC-DT-3: depends_on targets descendant

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/a/b"].
- Entry 3: logical_name = "ROOT/a/b".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### TC-DT-4: depends_on targets self

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ROOT/a"].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### TC-DT-5: depends_on with valid ROOT qualifier — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a".
- Entry 3: logical_name = "ROOT/b", frontmatter has depends_on = ["ROOT/a(interface)"].
  (Qualifier stripped before lookup: target = "ROOT/a", which exists and is
  neither ancestor, descendant, nor self of "ROOT/b".)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "dependency_targets".

---

#### TC-DT-6: depends_on with valid ARTIFACT reference — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "lib", path: "lib.go"}].
- Entry 3: logical_name = "ROOT/b", frontmatter has depends_on = ["ARTIFACT/a(lib)"].
  (Resolves to node "ROOT/a" with output id "lib", both of which exist.)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "dependency_targets".

---

#### TC-DT-7: depends_on with non-existent ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on = ["ARTIFACT/missing(id)"].
  (No entry with logical_name "ROOT/missing" exists.)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### TC-DT-8: Multiple invalid depends_on — one error per entry

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has depends_on =
  ["ROOT/missing", "ROOT/also_missing"].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least two FormatErrors, each with:
- node = "ROOT/a"
- rule = "dependency_targets"
One error per invalid dependency reference.

---

### Rule: input_target

---

#### TC-IT-1: Valid input reference — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "out", path: "a.go"}].
- Entry 3: logical_name = "ROOT/b", frontmatter has input = "ARTIFACT/a(out)".
  (Resolves to node "ROOT/a" with output id "out", both of which exist.)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2, Entry 3].

**Expected outcome:** No FormatError with rule = "input_target".

---

#### TC-IT-2: Input not starting with "ARTIFACT/"

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ROOT/something".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

#### TC-IT-3: Input references non-existent artifact

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has input = "ARTIFACT/missing(id)".
  (No entry with logical_name "ROOT/missing" exists.)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"

---

### Rule: external_files

---

#### TC-EF-1: External file exists — no fragments

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "some/file.txt"}].
- On disk: create file "some/file.txt" with any content (e.g., one line "hello").

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "external_files".

---

#### TC-EF-2: External file does not exist

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "nonexistent.txt"}].
- On disk: do not create "nonexistent.txt".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

#### TC-EF-3: Fragment with valid hash

**Setup:**
- On disk: create file "f.txt" with 5 lines:
  - Line 1: "alpha"
  - Line 2: "beta"
  - Line 3: "gamma"
  - Line 4: "delta"
  - Line 5: "epsilon"
- Compute the correct hash for lines 1–3:
  SHA-1 of the byte sequence "alpha\nbeta\ngamma\n",
  encoded as base64url (RFC 4648 §5, no padding) — always 27 characters.
  Call this value <correct-hash>.
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "1-3", hash: <correct-hash>}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "external_files".

---

#### TC-EF-4: Fragment with invalid hash

**Setup:**
- On disk: create file "f.txt" with 5 lines of known content (e.g., "alpha"
  through "epsilon" as in TC-EF-3).
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "1-3", hash: "wrong"}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

#### TC-EF-5: Fragment with invalid range format

**Setup:**
- On disk: create file "f.txt" with any content.
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "abc", hash: "x"}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

#### TC-EF-6: Fragment with start > end

**Setup:**
- On disk: create file "f.txt" with any content (at least 5 lines).
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "5-3", hash: "x"}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

#### TC-EF-7: Fragment with start < 1

**Setup:**
- On disk: create file "f.txt" with any content.
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "0-3", hash: "x"}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"

---

#### TC-EF-8: Fragment out of range

**Setup:**
- On disk: create file "f.txt" with exactly 5 lines of content.
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has external = [{path: "f.txt",
  fragments: [{lines: "1-100", hash: "x"}]}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates fragment is out of range

---

### Rule: output_paths

---

#### TC-OP-1: Valid output path — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x",
  path: "internal/x.go"}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "output_paths".

---

#### TC-OP-2: Output path with traversal

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x",
  path: "../../etc/passwd"}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

#### TC-OP-3: Output path with backslash

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", frontmatter has outputs = [{id: "x",
  path: "internal\\x.go"}].

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"

---

### Rule: duplicate_subsections

---

#### TC-DS-1: Unique subsection headings — no error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.public contains subsections with
  headings "Interface" and "Context" (both distinct).

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

#### TC-DS-2: Duplicate subsection headings — one error

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.public contains two subsections
  both with heading "Interface".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns exactly one FormatError where:
- node = "ROOT/a"
- rule = "duplicate_subsections"
(The second occurrence is flagged; the first occurrence is not.)

---

#### TC-DS-3: Three identical subsection headings — two errors

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.public contains three subsections
  all with heading "Interface".

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns exactly two FormatErrors, both with:
- node = "ROOT/a"
- rule = "duplicate_subsections"
(The second and third occurrences are flagged; the first is not.)

---

#### TC-DS-4: No public section — skip

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a", node.public is absent.

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** No FormatError with rule = "duplicate_subsections".

---

### Cross-Cutting

---

#### TC-CC-1: Collects multiple errors from different rules

**Setup:**
- Entry 1: logical_name = "ROOT".
- Entry 2: logical_name = "ROOT/a" with all of the following:
  - node.name_section.heading = "ROOT/wrong" (triggers name_heading)
  - frontmatter has depends_on = ["ROOT/missing"] (triggers dependency_targets)
  - node.public contains two subsections both named "Interface"
    (triggers duplicate_subsections)

**Action:** Call SpecTreeValidate with [Entry 1, Entry 2].

**Expected outcome:** Returns at least three FormatErrors covering at least three
distinct rules: "name_heading", "dependency_targets", and "duplicate_subsections".

---

#### TC-CC-2: Empty input list

**Setup:** No entries.

**Action:** Call SpecTreeValidate with an empty list.

**Expected outcome:** Returns an empty list of FormatErrors.
