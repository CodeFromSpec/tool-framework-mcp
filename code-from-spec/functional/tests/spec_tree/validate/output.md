<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@dvGSX8kCWtPILGncqu5iNFr1ttY -->

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
- Entry 1: logical_name = "ROOT", node has no agent section, node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a", frontmatter.depends_on = ["ROOT/b"] (where ROOT/b exists — see note), frontmatter.outputs = [{id: "out", path: "internal/a.go"}]

> Note: Add Entry 3: logical_name = "ROOT/b" to satisfy the dependency reference, and ensure ROOT is intermediate (ROOT/a exists as child). ROOT/a is a leaf (no entry starts with "ROOT/a/").

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns an empty list of FormatErrors.

---

### TC-HP-2: Valid intermediate node passes all checks

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT", no frontmatter fields (depends_on, outputs, input, external all absent), no agent section, node.public present with distinct subsections
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a", leaf node (no children)

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns an empty list of FormatErrors.

---

### TC-HP-3: Leaf with no frontmatter fields

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a", frontmatter is empty (no depends_on, no outputs, no input, no external), no agent section

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns an empty list of FormatErrors.

---

## Rule: name_heading

### TC-NH-1: Heading matches logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/a"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "name_heading".

---

### TC-NH-2: Heading does not match logical name

**Setup:**
- Entry 1: logical_name = "ROOT", node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.name_section.heading = "ROOT/wrong"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "name_heading"

---

## Rule: leaf_only_fields

### TC-LOF-1: Intermediate node with depends_on

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/b"] (ROOT/b exists as Entry 4)
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)
- Entry 4: logical_name = "ROOT/b"

**Actions:** Call SpecTreeValidate with all four entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- detail references "depends_on"

---

### TC-LOF-2: Intermediate node with outputs

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "x", path: "x.go"}]
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- detail references "outputs"

---

### TC-LOF-3: Intermediate node with input

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.input = "ARTIFACT/c(id)"
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- detail references "input"

---

### TC-LOF-4: Intermediate node with external

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "some/file.txt"}]
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- detail references "external"

---

### TC-LOF-5: Intermediate node with multiple restricted fields

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/b"] (ROOT/b exists as Entry 4), frontmatter.outputs = [{id: "x", path: "x.go"}]
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)
- Entry 4: logical_name = "ROOT/b"

**Actions:** Call SpecTreeValidate with all four entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "leaf_only_fields"
- One error references "depends_on", the other references "outputs"

---

## Rule: leaf_only_agent

### TC-LOA-1: Intermediate node with agent section

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.agent is present (non-absent)
- Entry 3: logical_name = "ROOT/a/b" (makes ROOT/a intermediate)

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "leaf_only_agent"

---

### TC-LOA-2: Leaf node with agent section — no error

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.agent is present, no children (ROOT/a is a leaf)

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "leaf_only_agent".

---

## Rule: dependency_targets

### TC-DT-1: depends_on targets non-existent ROOT node

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/missing"]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"
- detail references "ROOT/missing"

---

### TC-DT-2: depends_on targets ancestor

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a"
- Entry 3: logical_name = "ROOT/a/b", frontmatter.depends_on = ["ROOT"]

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a/b"
- rule = "dependency_targets"
- detail references "ROOT" as an ancestor

---

### TC-DT-3: depends_on targets descendant

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/a/b"]
- Entry 3: logical_name = "ROOT/a/b"

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"
- detail references "ROOT/a/b" as a descendant

---

### TC-DT-4: depends_on targets self

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/a"]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"
- detail references self-dependency

---

### TC-DT-5: depends_on with valid ROOT qualifier

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter.depends_on = ["ROOT/a(interface)"]

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns no FormatError with rule = "dependency_targets" (qualifier is stripped; "ROOT/a" exists and is not an ancestor, descendant, or self of "ROOT/b").

---

### TC-DT-6: depends_on with valid ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "lib", path: "lib.go"}]
- Entry 3: logical_name = "ROOT/b", frontmatter.depends_on = ["ARTIFACT/a(lib)"]

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns no FormatError with rule = "dependency_targets" (ARTIFACT reference resolves: logical_name "ROOT/a" exists and has output id "lib").

---

### TC-DT-7: depends_on with non-existent ARTIFACT reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ARTIFACT/missing(id)"]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "dependency_targets"
- detail references "ARTIFACT/missing(id)" as unresolvable

---

### TC-DT-8: Multiple invalid depends_on — one error per entry

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.depends_on = ["ROOT/missing", "ROOT/also_missing"]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "dependency_targets"
- One error references "ROOT/missing", the other references "ROOT/also_missing"

---

## Rule: input_target

### TC-IT-1: Valid input reference

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "out", path: "a.go"}]
- Entry 3: logical_name = "ROOT/b", frontmatter.input = "ARTIFACT/a(out)"

**Actions:** Call SpecTreeValidate with all three entries.

**Expected outcome:** Returns no FormatError with rule = "input_target".

---

### TC-IT-2: Input not starting with "ARTIFACT/"

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.input = "ROOT/something"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"
- detail indicates the input value does not start with "ARTIFACT/"

---

### TC-IT-3: Input references non-existent artifact

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.input = "ARTIFACT/missing(id)"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "input_target"
- detail references "ARTIFACT/missing(id)" as unresolvable

---

## Rule: external_files

### TC-EF-1: External file exists — no fragments

**Setup (disk):** Create file "some/file.txt" with any content.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "some/file.txt"}] (no fragments)

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "external_files".

---

### TC-EF-2: External file does not exist

**Setup (disk):** Do not create any file at "nonexistent.txt".

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "nonexistent.txt"}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail references "nonexistent.txt" as missing

---

### TC-EF-3: Fragment with valid hash

**Setup (disk):** Create file "f.txt" with the following 5 lines of known content:
```
line one
line two
line three
line four
line five
```
Compute the correct hash of lines 1–3 (i.e., "line one\nline two\nline three") and use it as the fragment hash.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "1-3", hash: <correct hash>}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "external_files".

---

### TC-EF-4: Fragment with invalid hash

**Setup (disk):** Create file "f.txt" with the following 5 lines:
```
line one
line two
line three
line four
line five
```

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "1-3", hash: "wrong"}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates hash mismatch for "f.txt" lines 1–3

---

### TC-EF-5: Fragment with invalid range format

**Setup (disk):** Create file "f.txt" with any content.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "abc", hash: "x"}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates "abc" is not a valid line range format

---

### TC-EF-6: Fragment with start > end

**Setup (disk):** Create file "f.txt" with any content.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "5-3", hash: "x"}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates start line (5) is greater than end line (3)

---

### TC-EF-7: Fragment with start < 1

**Setup (disk):** Create file "f.txt" with any content.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "0-3", hash: "x"}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates start line (0) is less than 1

---

### TC-EF-8: Fragment out of range

**Setup (disk):** Create file "f.txt" with exactly 5 lines of content.

**Setup (entries):**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.external = [{path: "f.txt", fragments: [{lines: "1-100", hash: "x"}]}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "external_files"
- detail indicates the fragment range (1–100) exceeds the file's total line count (5)

---

## Rule: output_paths

### TC-OP-1: Valid output path

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "x", path: "internal/x.go"}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "output_paths".

---

### TC-OP-2: Output path with traversal

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "x", path: "../../etc/passwd"}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"
- detail references the path "../../etc/passwd" as containing traversal sequences

---

### TC-OP-3: Output path with backslash

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter.outputs = [{id: "x", path: "internal\\x.go"}]

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least one FormatError where:
- node = "ROOT/a"
- rule = "output_paths"
- detail references the path "internal\\x.go" as containing a backslash

---

## Rule: duplicate_subsections

### TC-DS-1: Unique subsection headings — no error

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains subsections with headings "Interface" and "Context" (both distinct)

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "duplicate_subsections".

---

### TC-DS-2: Duplicate subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains two subsections both named "Interface"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly one FormatError where:
- node = "ROOT/a"
- rule = "duplicate_subsections"
- detail references the second occurrence of heading "Interface"

---

### TC-DS-3: Three identical subsection headings

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public contains three subsections all named "Interface"

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns exactly two FormatErrors where both have:
- node = "ROOT/a"
- rule = "duplicate_subsections"
- details reference the second and third occurrences of heading "Interface" respectively

---

### TC-DS-4: No public section — skip

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a", node.public is absent

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns no FormatError with rule = "duplicate_subsections".

---

## Cross-Cutting

### TC-CC-1: Collects multiple errors from different rules

**Setup:**
- Entry 1: logical_name = "ROOT"
- Entry 2: logical_name = "ROOT/a":
  - node.name_section.heading = "ROOT/wrong" (triggers name_heading)
  - frontmatter.depends_on = ["ROOT/missing"] (triggers dependency_targets)
  - node.public contains two subsections both named "Interface" (triggers duplicate_subsections)

**Actions:** Call SpecTreeValidate with both entries.

**Expected outcome:** Returns at least three FormatErrors, containing at least one each with rules "name_heading", "dependency_targets", and "duplicate_subsections", all with node = "ROOT/a".

---

### TC-CC-2: Empty input list

**Setup:** No entries.

**Actions:** Call SpecTreeValidate with an empty list.

**Expected outcome:** Returns an empty list of FormatErrors.
