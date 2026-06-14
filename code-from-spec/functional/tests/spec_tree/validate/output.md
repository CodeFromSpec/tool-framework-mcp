<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@bJE0y_TW38j3gG8grIJYTVlSPlE -->

# Test Specification: SpecTreeValidate

## Data Structures

`SpecTreeValidateInput` record fields:
- `logical_name`: string
- `frontmatter`: frontmatter.Frontmatter
- `node`: parsenode.Node

`FormatError` record fields:
- `node`: string
- `rule`: string
- `detail`: string

---

## Happy Path

### TC-HP-1: Valid leaf node passes all checks

Setup:
- Entry SPEC: intermediate node, node.name_section.heading = "spec", no frontmatter fields
- Entry SPEC/a: leaf, node.name_section.heading = "spec/a", depends_on = ["SPEC/b"], output = "internal/out.go"
- Entry SPEC/b: leaf, node.name_section.heading = "spec/b"
- all_dirs = ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Result is an empty list of FormatErrors

---

### TC-HP-2: Valid intermediate node passes all checks

Setup:
- Entry SPEC: intermediate, node.name_section.heading = "spec", node.public present with empty content, no frontmatter fields, no agent section
- Entry SPEC/a: leaf, node.name_section.heading = "spec/a"
- all_dirs = ["code-from-spec", "code-from-spec/a"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Result is an empty list of FormatErrors

---

### TC-HP-3: Leaf with no frontmatter fields

Setup:
- Entry SPEC: node.name_section.heading = "spec"
- Entry SPEC/a: leaf, node.name_section.heading = "spec/a", empty frontmatter
- all_dirs = ["code-from-spec", "code-from-spec/a"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Result is an empty list of FormatErrors

---

## name_heading

### TC-NH-1: Heading matches logical name

Setup:
- Entry SPEC: node.name_section.heading = "spec"
- Entry SPEC/a: node.name_section.heading = "spec/a"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "name_heading"

---

### TC-NH-2: Heading does not match logical name

Setup:
- Entry SPEC: node.name_section.heading = "spec"
- Entry SPEC/a: node.name_section.heading = "spec/wrong"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "name_heading" and node = "SPEC/a"

---

## leaf_only_fields

### TC-LOF-1: Intermediate node with depends_on

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), depends_on = ["SPEC/b"]
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "leaf_only_fields" and node = "SPEC/a"

---

### TC-LOF-2: Intermediate node with output

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), output = "x.go"
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "leaf_only_fields" and node = "SPEC/a"

---

### TC-LOF-3: Intermediate node with input

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), input = "ARTIFACT/c"
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "leaf_only_fields" and node = "SPEC/a"

---

### TC-LOF-4: Intermediate node with multiple restricted fields

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), depends_on = ["SPEC/b"], output = "x.go"
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly two FormatErrors, both with rule = "leaf_only_fields" and node = "SPEC/a" (one per restricted field)

---

## leaf_only_agent

### TC-LOA-1: Intermediate node with agent section

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), node.agent present with content = ["Agent instructions."]
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "leaf_only_agent" and node = "SPEC/a"

---

### TC-LOA-2: Leaf node with agent section — no error

Setup:
- Entry SPEC: node.name_section.heading = "spec"
- Entry SPEC/a: leaf, node.name_section.heading = "spec/a", node.agent present with content = ["Agent instructions."]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "leaf_only_agent"

---

## dependency_targets

### TC-DT-1: depends_on targets non-existent SPEC node

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["SPEC/missing"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-2: depends_on targets ancestor

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b)
- Entry SPEC/a/b: leaf, depends_on = ["SPEC"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a/b"

---

### TC-DT-3: depends_on targets descendant

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: intermediate (has child SPEC/a/b), depends_on = ["SPEC/a/b"]
- Entry SPEC/a/b: leaf

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-4: depends_on targets self

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["SPEC/a"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-5: depends_on with valid SPEC qualifier

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf
- Entry SPEC/b: leaf, depends_on = ["SPEC/a(interface)"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "dependency_targets"
  (qualifier "(interface)" is stripped before lookup, SPEC/a exists and is not ancestor/descendant/self of SPEC/b)

---

### TC-DT-6: depends_on with valid ARTIFACT reference

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, output = "lib.go"
- Entry SPEC/b: leaf, depends_on = ["ARTIFACT/a"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "dependency_targets"

---

### TC-DT-7: depends_on with non-existent ARTIFACT reference

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["ARTIFACT/missing"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-8: depends_on with valid EXTERNAL reference

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["EXTERNAL/proto/api.proto"]
- File on disk: "proto/api.proto" (any content)

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "dependency_targets"

---

### TC-DT-9: depends_on with non-existent EXTERNAL file

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["EXTERNAL/nonexistent.txt"]
- File "nonexistent.txt" does not exist on disk

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-10: depends_on with unrecognized prefix

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["UNKNOWN/something"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "dependency_targets" and node = "SPEC/a"

---

### TC-DT-11: Multiple invalid depends_on — one error per entry

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, depends_on = ["SPEC/missing", "SPEC/also_missing"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly two FormatErrors, both with rule = "dependency_targets" and node = "SPEC/a"

---

## input_target

### TC-IT-1: Valid ARTIFACT input reference

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, output = "a.go"
- Entry SPEC/b: leaf, input = "ARTIFACT/a"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "input_target"

---

### TC-IT-2: Valid EXTERNAL input reference

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, input = "EXTERNAL/docs/spec.yaml"
- File on disk: "docs/spec.yaml" (any content)

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "input_target"

---

### TC-IT-3: Input with unsupported prefix

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, input = "SPEC/something"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "input_target" and node = "SPEC/a"

---

### TC-IT-4: Input references non-existent artifact

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, input = "ARTIFACT/missing"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "input_target" and node = "SPEC/a"

---

### TC-IT-5: Input references non-existent EXTERNAL file

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, input = "EXTERNAL/nonexistent.txt"
- File "nonexistent.txt" does not exist on disk

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "input_target" and node = "SPEC/a"

---

## missing_node_md

### TC-MN-1: Subdirectory without _node.md

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf
- all_dirs = ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]
  (note: no entry exists for SPEC/b)

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "missing_node_md" and node = "code-from-spec/b"

---

### TC-MN-2: _-prefixed dir under code-from-spec — no error

Setup:
- Entry SPEC: intermediate
- all_dirs = ["code-from-spec", "code-from-spec/_rules"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "missing_node_md"
  (_-prefixed directories are ignored)

---

### TC-MN-3: All subdirectories have _node.md — no error

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf
- Entry SPEC/b: leaf
- all_dirs = ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "missing_node_md"

---

## output_paths

### TC-OP-1: Valid output path

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, output = "internal/x.go"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "output_paths"

---

### TC-OP-2: Output path with traversal

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, output = "../../etc/passwd"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "output_paths" and node = "SPEC/a"

---

### TC-OP-3: Output path with backslash

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, output = "internal\\x.go"

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "output_paths" and node = "SPEC/a"

---

## public_subsection_required

### TC-PSR-1: Public with content before first subsection

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with:
  - content = ["Some loose content."]
  - subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "public_subsection_required" and node = "SPEC/a"
  and detail = "content in # Public must be under a ## subsection"

---

### TC-PSR-2: Public with only blank lines before subsection — no error

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with:
  - content = ["", "  ", ""]
  - subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "public_subsection_required"

---

### TC-PSR-3: Public with content and no subsections

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with:
  - content = ["Some content."]
  - subsections = []

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "public_subsection_required" and node = "SPEC/a"

---

### TC-PSR-4: Public with only subsections — no error

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with:
  - content = []
  - subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "public_subsection_required"

---

### TC-PSR-5: No public section — skip

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public absent

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "public_subsection_required"

---

## duplicate_subsections

### TC-DS-1: Unique subsection headings — no error

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["Types."]}
  - {heading: "context", raw_heading: "## Context", content: ["Background."]}

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "duplicate_subsections"

---

### TC-DS-2: Duplicate subsection headings

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly one FormatError with rule = "duplicate_subsections" and node = "SPEC/a"
  (the second occurrence)

---

### TC-DS-3: Three identical subsection headings

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Third."]}

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Exactly two FormatErrors with rule = "duplicate_subsections" and node = "SPEC/a"
  (second and third occurrences)

---

### TC-DS-4: No public section — skip

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf, node.public absent

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- No FormatError with rule = "duplicate_subsections"

---

## Cross-cutting

### TC-CC-1: Collects multiple errors from different rules

Setup:
- Entry SPEC: intermediate
- Entry SPEC/a: leaf with:
  - node.name_section.heading = "spec/wrong"
  - depends_on = ["SPEC/missing"]
  - node.public present with subsections:
    - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
    - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- At least three FormatErrors for node = "SPEC/a":
  - One with rule = "name_heading"
  - One with rule = "dependency_targets"
  - One with rule = "duplicate_subsections"

---

### TC-CC-2: Empty input list

Setup:
- entries = empty list
- all_dirs = []

Actions:
- Call SpecTreeValidate(entries, all_dirs)

Expected outcome:
- Result is an empty list of FormatErrors
