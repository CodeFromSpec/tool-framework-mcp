<!-- code-from-spec: ROOT/functional/tests/spec_tree/validate@G1xGSsYFHGgr2AF4tdXkDQZOlUI -->

## Test cases for SpecTreeValidate

Each test case provides a list of `SpecTreeValidateInput` records, calls
`SpecTreeValidate`, and checks the returned list of `FormatError` records.

A `SpecTreeValidateInput` record has:
- `logical_name`: string
- `frontmatter`: a frontmatter record with optional fields `depends_on`,
  `output`, `input`, `external`
- `node`: a parsed node record with `name_section.heading`, optional
  `public` section, optional `agent` section

An entry is a leaf if no other entry has a logical name starting with
`<entry_logical_name>/`. An entry is intermediate otherwise.

---

### Happy path

#### Valid leaf node passes all checks

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields, node with heading = "ROOT",
  public present (no subsections), no agent section
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/b"],
  output = "internal/out.go", node with heading = "ROOT/a"
- logical_name = "ROOT/b", no frontmatter fields, node with heading = "ROOT/b"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns an empty list of `FormatError`.

---

#### Valid intermediate node passes all checks

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields, node with heading = "ROOT",
  public present with empty content, no agent section
- logical_name = "ROOT/a", no frontmatter fields, node with heading = "ROOT/a"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns an empty list of `FormatError`.

---

#### Leaf with no frontmatter fields

Setup: input list with two entries:
- logical_name = "ROOT", node with heading = "ROOT"
- logical_name = "ROOT/a", empty frontmatter, node with heading = "ROOT/a"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns an empty list of `FormatError`.

---

### name_heading

#### Heading matches logical name

Setup: input list with two entries:
- logical_name = "ROOT", node with heading = "ROOT"
- logical_name = "ROOT/a", node with heading = "ROOT/a"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "name_heading".

---

#### Heading does not match logical name

Setup: input list with two entries:
- logical_name = "ROOT", node with heading = "ROOT"
- logical_name = "ROOT/a", node with heading = "ROOT/wrong"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "name_heading"

---

### leaf_only_fields

#### Intermediate node with depends_on

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/b"],
  node with heading = "ROOT/a"
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### Intermediate node with output

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "x.go",
  node with heading = "ROOT/a"
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### Intermediate node with input

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter input = "ARTIFACT/c",
  node with heading = "ROOT/a"
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### Intermediate node with external

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter external = [{path: "some/file.txt"}],
  node with heading = "ROOT/a"
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "leaf_only_fields"

---

#### Intermediate node with multiple restricted fields

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/b"],
  output = "x.go", node with heading = "ROOT/a"
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly two `FormatError` records, both with:
- node = "ROOT/a"
- rule = "leaf_only_fields"
(one per restricted field that is set)

---

### leaf_only_agent

#### Intermediate node with agent section

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", no frontmatter fields, node with heading = "ROOT/a",
  agent section present with content = ["Agent instructions."]
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "leaf_only_agent"

---

#### Leaf node with agent section — no error

Setup: input list with two entries:
- logical_name = "ROOT", node with heading = "ROOT"
- logical_name = "ROOT/a", node with heading = "ROOT/a",
  agent section present with content = ["Agent instructions."]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "leaf_only_agent".

---

### dependency_targets

#### depends_on targets non-existent ROOT node

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/missing"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### depends_on targets ancestor

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", no frontmatter fields
- logical_name = "ROOT/a/b", frontmatter depends_on = ["ROOT"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a/b"
- rule = "dependency_targets"

---

#### depends_on targets descendant

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/a/b"]
- logical_name = "ROOT/a/b", no frontmatter fields

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### depends_on targets self

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/a"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### depends_on with valid ROOT qualifier

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", no frontmatter fields
- logical_name = "ROOT/b", frontmatter depends_on = ["ROOT/a(interface)"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "dependency_targets".
(The qualifier "(interface)" is stripped before lookup; "ROOT/a" exists and
is not an ancestor, descendant, or self relative to "ROOT/b".)

---

#### depends_on with valid ARTIFACT reference

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "lib.go"
- logical_name = "ROOT/b", frontmatter depends_on = ["ARTIFACT/a"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "dependency_targets".

---

#### depends_on with non-existent ARTIFACT reference

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ARTIFACT/missing"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "dependency_targets"

---

#### Multiple invalid depends_on — one error per entry

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter depends_on = ["ROOT/missing", "ROOT/also_missing"]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly two `FormatError` records, both with:
- node = "ROOT/a"
- rule = "dependency_targets"
(one per invalid dependency reference)

---

### input_target

#### Valid input reference

Setup: input list with three entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "a.go"
- logical_name = "ROOT/b", frontmatter input = "ARTIFACT/a"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "input_target".

---

#### Input not starting with ARTIFACT/

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter input = "ROOT/something"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "input_target"

---

#### Input references non-existent artifact

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter input = "ARTIFACT/missing"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "input_target"

---

### external_files

#### External file exists

Setup: create a file at path "some/file.txt" on disk with content "hello\n".
Input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter external = [{path: "some/file.txt"}]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "external_files".

---

#### External file does not exist

Setup: do not create any file at "nonexistent.txt".
Input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter external = [{path: "nonexistent.txt"}]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "external_files"

---

### output_paths

#### Valid output path

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "internal/x.go"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "output_paths".

---

#### Output path with traversal

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "../../etc/passwd"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "output_paths"

---

#### Output path with backslash

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", frontmatter output = "internal\\x.go"

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "output_paths"

---

### public_subsection_required

#### Public with content before first subsection

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section present,
  content = ["Some loose content."],
  subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "public_subsection_required"
- detail = "content in # Public must be under a ## subsection"

---

#### Public with only blank lines before subsection — no error

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section present,
  content = ["", "  ", ""],
  subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "public_subsection_required".

---

#### Public with content and no subsections

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section present,
  content = ["Some content."], subsections = []

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "public_subsection_required"

---

#### Public with only subsections — no error

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section present,
  content = [],
  subsections = [{heading: "interface", raw_heading: "## Interface", content: ["Types."]}]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "public_subsection_required".

---

#### No public section — skip

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section absent

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "public_subsection_required".

---

### duplicate_subsections

#### Unique subsection headings — no error

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section containing subsections:
  - heading = "interface", raw_heading = "## Interface", content = ["Types."]
  - heading = "context", raw_heading = "## Context", content = ["Background."]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "duplicate_subsections".

---

#### Duplicate subsection headings

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section containing subsections:
  - heading = "interface", raw_heading = "## Interface", content = ["First."]
  - heading = "interface", raw_heading = "## Interface", content = ["Second."]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly one `FormatError` with:
- node = "ROOT/a"
- rule = "duplicate_subsections"
(for the second occurrence of heading "interface")

---

#### Three identical subsection headings

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section containing subsections:
  - heading = "interface", raw_heading = "## Interface", content = ["First."]
  - heading = "interface", raw_heading = "## Interface", content = ["Second."]
  - heading = "interface", raw_heading = "## Interface", content = ["Third."]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns exactly two `FormatError` records, both with:
- node = "ROOT/a"
- rule = "duplicate_subsections"
(for the second and third occurrences of heading "interface")

---

#### No public section — skip

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with public section absent

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: returns no `FormatError` with rule = "duplicate_subsections".

---

### Cross-cutting

#### Collects multiple errors from different rules

Setup: input list with two entries:
- logical_name = "ROOT", no frontmatter fields
- logical_name = "ROOT/a", node with heading = "ROOT/wrong",
  frontmatter depends_on = ["ROOT/missing"],
  public section containing subsections:
  - heading = "interface", raw_heading = "## Interface", content = ["First."]
  - heading = "interface", raw_heading = "## Interface", content = ["Second."]

Actions: call `SpecTreeValidate` with the input list.

Expected outcome: the returned list contains at least three `FormatError` records:
- one with rule = "name_heading" for node "ROOT/a"
- one with rule = "dependency_targets" for node "ROOT/a"
- one with rule = "duplicate_subsections" for node "ROOT/a"

---

#### Empty input list

Setup: empty input list.

Actions: call `SpecTreeValidate` with the empty list.

Expected outcome: returns an empty list of `FormatError`.
