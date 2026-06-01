<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@NZpMWPPQ-vXSkIx52QXud2-lCGs -->

## Test Cases for SpecTreeValidate

---

### Happy path

#### Valid leaf node passes all checks

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT", node has children ROOT/a and ROOT/b
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/b"], outputs = [{id: "out", path: "internal/out.go"}]), node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter = (empty), node.name_section.heading = "ROOT/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns empty list of FormatErrors.

---

#### Valid intermediate node passes all checks

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT", node.public present with empty content, no agent section
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns empty list of FormatErrors.

---

#### Leaf with no frontmatter fields

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns empty list of FormatErrors.

---

### name_heading

#### Heading matches logical name

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "name_heading".

---

#### Heading does not match logical name

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/wrong"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "name_heading".

---

### leaf_only_fields

#### Intermediate node with depends_on

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/b"]), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "leaf_only_fields".

---

#### Intermediate node with outputs

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "x", path: "x.go"}]), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "leaf_only_fields".

---

#### Intermediate node with input

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (input = "ARTIFACT/c(id)"), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "leaf_only_fields".

---

#### Intermediate node with external

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (external = [{path: "some/file.txt"}]), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "leaf_only_fields".

---

#### Intermediate node with multiple restricted fields

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/b"], outputs = [{id: "x", path: "x.go"}]), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns exactly two FormatErrors where node = "ROOT/a" and rule = "leaf_only_fields" (one per restricted field present).

---

### leaf_only_agent

#### Intermediate node with agent section

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b), node.agent present with content = ["Agent instructions."]
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "leaf_only_agent".

---

#### Leaf node with agent section — no error

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a", node.agent present with content = ["Agent instructions."]

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "leaf_only_agent".

---

### dependency_targets

#### depends_on targets non-existent ROOT node

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/missing"]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "dependency_targets".

---

#### depends_on targets ancestor

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (depends_on = ["ROOT"]), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a/b" and rule = "dependency_targets".

---

#### depends_on targets descendant

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/a/b"]), node.name_section.heading = "ROOT/a" (intermediate — has child ROOT/a/b)
- Entry 3: logical_name = "ROOT/a/b", frontmatter = (empty), node.name_section.heading = "ROOT/a/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "dependency_targets".

---

#### depends_on targets self

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/a"]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "dependency_targets".

---

#### depends_on with valid ROOT qualifier

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter = (depends_on = ["ROOT/a(interface)"]), node.name_section.heading = "ROOT/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns no FormatError with rule = "dependency_targets". (Qualifier is stripped before lookup; ROOT/a exists and is neither ancestor, descendant, nor self of ROOT/b.)

---

#### depends_on with valid ARTIFACT reference

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "lib", path: "lib.go"}]), node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter = (depends_on = ["ARTIFACT/a(lib)"]), node.name_section.heading = "ROOT/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns no FormatError with rule = "dependency_targets".

---

#### depends_on with non-existent ARTIFACT reference

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ARTIFACT/missing(id)"]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "dependency_targets".

---

#### Multiple invalid depends_on — one error per entry

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/missing", "ROOT/also_missing"]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns exactly two FormatErrors where node = "ROOT/a" and rule = "dependency_targets" (one per invalid entry).

---

### input_target

#### Valid input reference

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "out", path: "a.go"}]), node.name_section.heading = "ROOT/a"
- Entry 3: logical_name = "ROOT/b", frontmatter = (input = "ARTIFACT/a(out)"), node.name_section.heading = "ROOT/b"

Actions: Call SpecTreeValidate with the three entries.

Expected outcome: Returns no FormatError with rule = "input_target".

---

#### Input not starting with ARTIFACT/

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (input = "ROOT/something"), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "input_target".

---

#### Input references non-existent artifact

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (input = "ARTIFACT/missing(id)"), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "input_target".

---

### external_files

#### External file exists

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (external = [{path: "some/file.txt"}]), node.name_section.heading = "ROOT/a"
- File on disk: create "some/file.txt" with content "hello\n" before calling the function.

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "external_files".

---

#### External file does not exist

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (external = [{path: "nonexistent.txt"}]), node.name_section.heading = "ROOT/a"
- File on disk: do not create "nonexistent.txt".

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "external_files".

---

### output_paths

#### Valid output path

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "x", path: "internal/x.go"}]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "output_paths".

---

#### Output path with traversal

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "x", path: "../../etc/passwd"}]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "output_paths".

---

#### Output path with backslash

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (outputs = [{id: "x", path: "internal\\x.go"}]), node.name_section.heading = "ROOT/a"

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least one FormatError where node = "ROOT/a" and rule = "output_paths".

---

### duplicate_subsections

#### Unique subsection headings — no error

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a", node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["Types."]}
  - {heading: "context", raw_heading: "## Context", content: ["Background."]}

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "duplicate_subsections".

---

#### Duplicate subsection headings

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a", node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns exactly one FormatError where node = "ROOT/a" and rule = "duplicate_subsections" (for the second occurrence).

---

#### Three identical subsection headings

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a", node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Third."]}

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns exactly two FormatErrors where node = "ROOT/a" and rule = "duplicate_subsections" (for the second and third occurrences).

---

#### No public section — skip

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (empty), node.name_section.heading = "ROOT/a", node.public absent

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns no FormatError with rule = "duplicate_subsections".

---

### Cross-cutting

#### Collects multiple errors from different rules

Setup:
- Entry 1: logical_name = "ROOT", frontmatter = (empty), node.name_section.heading = "ROOT"
- Entry 2: logical_name = "ROOT/a", frontmatter = (depends_on = ["ROOT/missing"]), node.name_section.heading = "ROOT/wrong", node.public present with subsections:
  - {heading: "interface", raw_heading: "## Interface", content: ["First."]}
  - {heading: "interface", raw_heading: "## Interface", content: ["Second."]}

Actions: Call SpecTreeValidate with the two entries.

Expected outcome: Returns at least three FormatErrors for node = "ROOT/a":
- At least one with rule = "name_heading"
- At least one with rule = "dependency_targets"
- At least one with rule = "duplicate_subsections"

---

#### Empty input list

Setup: No entries.

Actions: Call SpecTreeValidate with an empty list.

Expected outcome: Returns empty list of FormatErrors.
