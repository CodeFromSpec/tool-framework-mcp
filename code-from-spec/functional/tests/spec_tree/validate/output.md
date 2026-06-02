<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@4Ei9QqX3_HKD9POSqHGpploEJgY -->

# Test cases for SpecTreeValidate

Input is always a list of `SpecTreeValidateInput`. Each entry has: logical_name,
frontmatter (Frontmatter record), and node (Node record). Except for external_files
tests, no file I/O is performed.

---

## Happy path

### Valid leaf node passes all checks

Setup: input list =
- ROOT: intermediate (has children ROOT/a and ROOT/b),
  node.name_section.heading = "ROOT", no frontmatter fields, no agent section
- ROOT/a: leaf, node.name_section.heading = "ROOT/a",
  depends_on = ["ROOT/b"], output = "internal/out.go", node.public present
- ROOT/b: leaf, node.name_section.heading = "ROOT/b"

Action: call SpecTreeValidate.

Expect: no format errors.

---

### Valid intermediate node passes all checks

Setup: input list =
- ROOT: intermediate (has child ROOT/a), node.name_section.heading = "ROOT",
  node.public present with empty content, no frontmatter fields, no agent section
- ROOT/a: leaf, node.name_section.heading = "ROOT/a"

Action: call SpecTreeValidate.

Expect: no format errors.

---

### Leaf with no frontmatter fields

Setup: input list =
- ROOT: node.name_section.heading = "ROOT"
- ROOT/a: leaf, node.name_section.heading = "ROOT/a", empty frontmatter

Action: call SpecTreeValidate.

Expect: no format errors.

---

## Rule: name_heading

### Heading matches logical name

Setup: input list =
- ROOT: node.name_section.heading = "ROOT"
- ROOT/a: node.name_section.heading = "ROOT/a"

Action: call SpecTreeValidate.

Expect: no format error with rule = "name_heading".

---

### Heading does not match logical name

Setup: input list =
- ROOT: node.name_section.heading = "ROOT"
- ROOT/a: node.name_section.heading = "ROOT/wrong"

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "name_heading".

---

## Rule: leaf_only_fields

### Intermediate node with depends_on

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), depends_on = ["ROOT/b"]
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "leaf_only_fields".

---

### Intermediate node with output

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), output = "x.go"
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "leaf_only_fields".

---

### Intermediate node with input

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), input = "ARTIFACT/c"
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "leaf_only_fields".

---

### Intermediate node with external

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), external = [{path: "some/file.txt"}]
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "leaf_only_fields".

---

### Intermediate node with multiple restricted fields

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), depends_on = ["ROOT/b"], output = "x.go"
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: two FormatErrors both with node = "ROOT/a", rule = "leaf_only_fields"
(one per restricted field).

---

## Rule: leaf_only_agent

### Intermediate node with agent section

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), node.agent present with
  content = ["Agent instructions."]
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "leaf_only_agent".

---

### Leaf node with agent section — no error

Setup: input list =
- ROOT: node.name_section.heading = "ROOT"
- ROOT/a: leaf, node.name_section.heading = "ROOT/a",
  node.agent present with content = ["Agent instructions."]

Action: call SpecTreeValidate.

Expect: no format error with rule = "leaf_only_agent".

---

## Rule: dependency_targets

### depends_on targets non-existent ROOT node

Setup: input list =
- ROOT
- ROOT/a: leaf, depends_on = ["ROOT/missing"]

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "dependency_targets".

---

### depends_on targets ancestor

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b)
- ROOT/a/b: leaf, depends_on = ["ROOT"]

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a/b", rule = "dependency_targets".

---

### depends_on targets descendant

Setup: input list =
- ROOT
- ROOT/a: intermediate (has child ROOT/a/b), depends_on = ["ROOT/a/b"]
- ROOT/a/b: leaf

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "dependency_targets".

---

### depends_on targets self

Setup: input list =
- ROOT
- ROOT/a: leaf, depends_on = ["ROOT/a"]

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "dependency_targets".

---

### depends_on with valid ROOT qualifier

Setup: input list =
- ROOT
- ROOT/a: leaf
- ROOT/b: leaf, depends_on = ["ROOT/a(interface)"]

Action: call SpecTreeValidate.

Expect: no format error with rule = "dependency_targets"
(qualifier stripped, ROOT/a exists and is not ancestor/descendant/self of ROOT/b).

---

### depends_on with valid ARTIFACT reference

Setup: input list =
- ROOT
- ROOT/a: leaf, output = "lib.go"
- ROOT/b: leaf, depends_on = ["ARTIFACT/a"]

Action: call SpecTreeValidate.

Expect: no format error with rule = "dependency_targets".

---

### depends_on with non-existent ARTIFACT reference

Setup: input list =
- ROOT
- ROOT/a: leaf, depends_on = ["ARTIFACT/missing"]

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "dependency_targets".

---

### Multiple invalid depends_on — one error per entry

Setup: input list =
- ROOT
- ROOT/a: leaf, depends_on = ["ROOT/missing", "ROOT/also_missing"]

Action: call SpecTreeValidate.

Expect: two FormatErrors both with node = "ROOT/a", rule = "dependency_targets".

---

## Rule: input_target

### Valid input reference

Setup: input list =
- ROOT
- ROOT/a: leaf, output = "a.go"
- ROOT/b: leaf, input = "ARTIFACT/a"

Action: call SpecTreeValidate.

Expect: no format error with rule = "input_target".

---

### Input not starting with ARTIFACT/

Setup: input list =
- ROOT
- ROOT/a: leaf, input = "ROOT/something"

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "input_target".

---

### Input references non-existent artifact

Setup: input list =
- ROOT
- ROOT/a: leaf, input = "ARTIFACT/missing"

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "input_target".

---

## Rule: external_files

### External file exists

Setup: input list =
- ROOT
- ROOT/a: leaf, external = [{path: "some/file.txt"}]

Create "some/file.txt" on disk with content "hello\n".

Action: call SpecTreeValidate.

Expect: no format error with rule = "external_files".

---

### External file does not exist

Setup: input list =
- ROOT
- ROOT/a: leaf, external = [{path: "nonexistent.txt"}]

Do not create "nonexistent.txt" on disk.

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "external_files".

---

## Rule: output_paths

### Valid output path

Setup: input list =
- ROOT
- ROOT/a: leaf, output = "internal/x.go"

Action: call SpecTreeValidate.

Expect: no format error with rule = "output_paths".

---

### Output path with traversal

Setup: input list =
- ROOT
- ROOT/a: leaf, output = "../../etc/passwd"

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "output_paths".

---

### Output path with backslash

Setup: input list =
- ROOT
- ROOT/a: leaf, output = "internal\\x.go"

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "output_paths".

---

## Rule: duplicate_subsections

### Unique subsection headings — no error

Setup: input list =
- ROOT
- ROOT/a: leaf, node.public present with subsections:
    [{heading: "interface", raw_heading: "## Interface", content: ["Types."]},
     {heading: "context", raw_heading: "## Context", content: ["Background."]}]

Action: call SpecTreeValidate.

Expect: no format error with rule = "duplicate_subsections".

---

### Duplicate subsection headings

Setup: input list =
- ROOT
- ROOT/a: leaf, node.public present with subsections:
    [{heading: "interface", raw_heading: "## Interface", content: ["First."]},
     {heading: "interface", raw_heading: "## Interface", content: ["Second."]}]

Action: call SpecTreeValidate.

Expect: one FormatError with node = "ROOT/a", rule = "duplicate_subsections"
(for the second occurrence).

---

### Three identical subsection headings

Setup: input list =
- ROOT
- ROOT/a: leaf, node.public present with subsections:
    [{heading: "interface", raw_heading: "## Interface", content: ["First."]},
     {heading: "interface", raw_heading: "## Interface", content: ["Second."]},
     {heading: "interface", raw_heading: "## Interface", content: ["Third."]}]

Action: call SpecTreeValidate.

Expect: two FormatErrors both with node = "ROOT/a", rule = "duplicate_subsections"
(for the second and third occurrences).

---

### No public section — skip

Setup: input list =
- ROOT
- ROOT/a: leaf, node.public absent

Action: call SpecTreeValidate.

Expect: no format error with rule = "duplicate_subsections".

---

## Cross-cutting

### Collects multiple errors from different rules

Setup: input list =
- ROOT
- ROOT/a: leaf, node.name_section.heading = "ROOT/wrong",
  depends_on = ["ROOT/missing"],
  node.public present with subsections:
    [{heading: "interface", raw_heading: "## Interface", content: ["First."]},
     {heading: "interface", raw_heading: "## Interface", content: ["Second."]}]

Action: call SpecTreeValidate.

Expect: at least three FormatErrors for ROOT/a:
- one with rule = "name_heading"
- one with rule = "dependency_targets"
- one with rule = "duplicate_subsections"

---

### Empty input list

Setup: input = empty list.

Action: call SpecTreeValidate.

Expect: no format errors (empty list returned).
