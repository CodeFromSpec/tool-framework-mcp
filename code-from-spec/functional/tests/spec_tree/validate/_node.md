---
depends_on:
  - ROOT/functional/logic/spec_tree/validate(interface)
output: code-from-spec/functional/tests/spec_tree/validate/output.md
---

# ROOT/functional/tests/spec_tree/validate

Test cases for the spec tree validation component.

# Public

## Test cases

### Happy path

#### Valid leaf node passes all checks

Input: SPEC (intermediate, has children SPEC/a and
SPEC/b), SPEC/a (leaf, node.name_section.heading =
"spec/a", depends_on = ["SPEC/b"], output =
"internal/out.go"), SPEC/b (leaf,
node.name_section.heading = "spec/b"). all_dirs =
["code-from-spec", "code-from-spec/a",
"code-from-spec/b"]. Call SpecTreeValidate. Expect no
format errors.

#### Valid intermediate node passes all checks

Input: SPEC (intermediate, node.name_section.heading =
"spec", node.public present with empty content, no
frontmatter fields, no agent section), SPEC/a (leaf,
node.name_section.heading = "spec/a"). all_dirs =
["code-from-spec", "code-from-spec/a"]. Call
SpecTreeValidate. Expect no format errors.

#### Leaf with no frontmatter fields

Input: SPEC (node.name_section.heading = "spec"),
SPEC/a (leaf, node.name_section.heading = "spec/a",
empty frontmatter). all_dirs = ["code-from-spec",
"code-from-spec/a"]. Call SpecTreeValidate. Expect no
format errors.

### name_heading

#### Heading matches logical name

Input: SPEC (node.name_section.heading = "spec"),
SPEC/a (node.name_section.heading = "spec/a"). Call
SpecTreeValidate. Expect no name_heading error.

#### ROOT/ heading matches SPEC/ logical name

Input: SPEC (node.name_section.heading = "root"),
SPEC/a (node.name_section.heading = "root/a"). Call
SpecTreeValidate. Expect no name_heading error —
root/ prefix is treated as alias for spec/.

#### Heading does not match logical name

Input: SPEC (node.name_section.heading = "spec"),
SPEC/a (node.name_section.heading = "spec/wrong").
Call SpecTreeValidate. Expect a FormatError with rule
= "name_heading" for SPEC/a.

### leaf_only_fields

#### Intermediate node with depends_on

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, depends_on = ["SPEC/b"]), SPEC/a/b (leaf).
Call SpecTreeValidate. Expect a FormatError with rule
= "leaf_only_fields" for SPEC/a.

#### Intermediate node with output

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, output = "x.go"),
SPEC/a/b (leaf). Call SpecTreeValidate. Expect a
FormatError with rule = "leaf_only_fields" for SPEC/a.

#### Intermediate node with input

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, input = "ARTIFACT/c"), SPEC/a/b (leaf).
Call SpecTreeValidate. Expect a FormatError with rule
= "leaf_only_fields" for SPEC/a.

#### Intermediate node with multiple restricted fields

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, depends_on = ["SPEC/b"], output = "x.go"),
SPEC/a/b (leaf). Call SpecTreeValidate. Expect two
FormatErrors with rule = "leaf_only_fields" for SPEC/a
(one per field).

### leaf_only_agent

#### Intermediate node with agent section

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, node.agent present with content =
["Agent instructions."]), SPEC/a/b (leaf). Call
SpecTreeValidate. Expect a FormatError with rule =
"leaf_only_agent" for SPEC/a.

#### Leaf node with agent section — no error

Input: SPEC (node.name_section.heading = "spec"),
SPEC/a (leaf, node.name_section.heading = "spec/a",
node.agent present with content = ["Agent
instructions."]). Call SpecTreeValidate. Expect no
leaf_only_agent error.

### dependency_targets

#### depends_on targets non-existent SPEC node

Input: SPEC, SPEC/a (leaf, depends_on =
["SPEC/missing"]). Call SpecTreeValidate. Expect a
FormatError with rule = "dependency_targets" for
SPEC/a.

#### depends_on with ROOT/ reference normalized

Input: SPEC, SPEC/a (leaf), SPEC/b (leaf, depends_on
= ["ROOT/a"]). Call SpecTreeValidate. Expect no
dependency_targets error — ROOT/a is normalized to
SPEC/a, which exists.

#### depends_on targets ancestor

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b), SPEC/a/b (leaf, depends_on = ["SPEC"]).
Call SpecTreeValidate. Expect a FormatError with rule
= "dependency_targets" for SPEC/a/b.

#### depends_on targets descendant

Input: SPEC, SPEC/a (intermediate, has child
SPEC/a/b, depends_on = ["SPEC/a/b"]), SPEC/a/b
(leaf). Call SpecTreeValidate. Expect a FormatError
with rule = "dependency_targets" for SPEC/a.

#### depends_on targets self

Input: SPEC, SPEC/a (leaf, depends_on = ["SPEC/a"]).
Call SpecTreeValidate. Expect a FormatError with rule
= "dependency_targets" for SPEC/a.

#### depends_on with valid SPEC qualifier

Input: SPEC, SPEC/a (leaf), SPEC/b (leaf, depends_on
= ["SPEC/a(interface)"]). Call SpecTreeValidate.
Expect no dependency_targets error (qualifier
stripped, SPEC/a exists and is not
ancestor/descendant/self).

#### depends_on with valid ARTIFACT reference

Input: SPEC, SPEC/a (leaf, output = "lib.go"), SPEC/b
(leaf, depends_on = ["ARTIFACT/a"]). Call
SpecTreeValidate. Expect no dependency_targets error.

#### depends_on with non-existent ARTIFACT reference

Input: SPEC, SPEC/a (leaf, depends_on =
["ARTIFACT/missing"]). Call SpecTreeValidate.
Expect a FormatError with rule = "dependency_targets"
for SPEC/a.

#### depends_on with valid EXTERNAL reference

Input: SPEC, SPEC/a (leaf, depends_on =
["EXTERNAL/proto/api.proto"]). Create
"proto/api.proto" on disk. Call SpecTreeValidate.
Expect no dependency_targets error.

#### depends_on with non-existent EXTERNAL file

Input: SPEC, SPEC/a (leaf, depends_on =
["EXTERNAL/nonexistent.txt"]). Do not create the file.
Call SpecTreeValidate. Expect a FormatError with rule
= "dependency_targets" for SPEC/a.

#### depends_on with unrecognized prefix

Input: SPEC, SPEC/a (leaf, depends_on =
["UNKNOWN/something"]). Call SpecTreeValidate. Expect
a FormatError with rule = "dependency_targets" for
SPEC/a.

#### Multiple invalid depends_on — one error per entry

Input: SPEC, SPEC/a (leaf, depends_on =
["SPEC/missing", "SPEC/also_missing"]). Call
SpecTreeValidate. Expect two FormatErrors with rule =
"dependency_targets" for SPEC/a.

### input_target

#### Valid ARTIFACT input reference

Input: SPEC, SPEC/a (leaf, output = "a.go"), SPEC/b
(leaf, input = "ARTIFACT/a"). Call SpecTreeValidate.
Expect no input_target error.

#### Valid EXTERNAL input reference

Input: SPEC, SPEC/a (leaf, input =
"EXTERNAL/docs/spec.yaml"). Create
"docs/spec.yaml" on disk. Call SpecTreeValidate.
Expect no input_target error.

#### Input with unsupported prefix

Input: SPEC, SPEC/a (leaf, input = "SPEC/something").
Call SpecTreeValidate. Expect a FormatError with rule
= "input_target" for SPEC/a.

#### Input references non-existent artifact

Input: SPEC, SPEC/a (leaf, input =
"ARTIFACT/missing"). Call SpecTreeValidate.
Expect a FormatError with rule = "input_target" for
SPEC/a.

#### Input references non-existent EXTERNAL file

Input: SPEC, SPEC/a (leaf, input =
"EXTERNAL/nonexistent.txt"). Do not create the file.
Call SpecTreeValidate. Expect a FormatError with rule
= "input_target" for SPEC/a.

### missing_node_md

#### Subdirectory without _node.md

Input: SPEC, SPEC/a (leaf). all_dirs =
["code-from-spec", "code-from-spec/a",
"code-from-spec/b"]. Call SpecTreeValidate. Expect a
FormatError with rule = "missing_node_md" for
directory "code-from-spec/b".

#### _-prefixed dir under code-from-spec — no error

Input: SPEC. all_dirs = ["code-from-spec",
"code-from-spec/_rules"]. Call SpecTreeValidate.
Expect no missing_node_md error — _-prefixed
directories are ignored.

#### All subdirectories have _node.md — no error

Input: SPEC, SPEC/a (leaf), SPEC/b (leaf). all_dirs =
["code-from-spec", "code-from-spec/a",
"code-from-spec/b"]. Call SpecTreeValidate. Expect no
missing_node_md error.

### output_paths

#### Valid output path

Input: SPEC, SPEC/a (leaf, output = "internal/x.go").
Call SpecTreeValidate. Expect no output_paths error.

#### Output path with traversal

Input: SPEC, SPEC/a (leaf, output =
"../../etc/passwd"). Call SpecTreeValidate. Expect a
FormatError with rule = "output_paths" for SPEC/a.

#### Output path with backslash

Input: SPEC, SPEC/a (leaf, output =
"internal\\x.go"). Call SpecTreeValidate. Expect a
FormatError with rule = "output_paths" for SPEC/a.

### public_subsection_required

#### Public with content before first subsection

Input: SPEC, SPEC/a (leaf, node.public present with
content = ["Some loose content."], subsections =
[{heading: "interface", raw_heading: "## Interface",
content: ["Types."]}]). Call SpecTreeValidate. Expect
a FormatError with rule = "public_subsection_required"
for SPEC/a with detail "content in # Public must be
under a ## subsection".

#### Public with only blank lines before subsection — no error

Input: SPEC, SPEC/a (leaf, node.public present with
content = ["", "  ", ""], subsections =
[{heading: "interface", raw_heading: "## Interface",
content: ["Types."]}]). Call SpecTreeValidate. Expect
no public_subsection_required error.

#### Public with content and no subsections

Input: SPEC, SPEC/a (leaf, node.public present with
content = ["Some content."], subsections = []). Call
SpecTreeValidate. Expect a FormatError with rule =
"public_subsection_required" for SPEC/a.

#### Public with only subsections — no error

Input: SPEC, SPEC/a (leaf, node.public present with
content = [], subsections = [{heading: "interface",
raw_heading: "## Interface", content: ["Types."]}]).
Call SpecTreeValidate. Expect no
public_subsection_required error.

#### No public section — skip

Input: SPEC, SPEC/a (leaf, node.public absent). Call
SpecTreeValidate. Expect no public_subsection_required
error.

### duplicate_subsections

#### Unique subsection headings — no error

Input: SPEC, SPEC/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["Types."]}, {heading:
"context", raw_heading: "## Context", content:
["Background."]}]). Call SpecTreeValidate. Expect no
duplicate_subsections error.

#### Duplicate subsection headings

Input: SPEC, SPEC/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["First."]}, {heading:
"interface", raw_heading: "## Interface", content:
["Second."]}]). Call SpecTreeValidate. Expect one
FormatError with rule = "duplicate_subsections" for
SPEC/a (the second occurrence).

#### Three identical subsection headings

Input: SPEC, SPEC/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["First."]}, {heading:
"interface", raw_heading: "## Interface", content:
["Second."]}, {heading: "interface", raw_heading:
"## Interface", content: ["Third."]}]). Call
SpecTreeValidate. Expect two FormatErrors with rule =
"duplicate_subsections" for SPEC/a (second and third
occurrences).

#### No public section — skip

Input: SPEC, SPEC/a (leaf, node.public absent). Call
SpecTreeValidate. Expect no duplicate_subsections
error.

### Cross-cutting

#### Collects multiple errors from different rules

Input: SPEC, SPEC/a (leaf, node.name_section.heading
= "spec/wrong", depends_on = ["SPEC/missing"],
node.public present with subsections [{heading:
"interface", raw_heading: "## Interface", content:
["First."]}, {heading: "interface", raw_heading:
"## Interface", content: ["Second."]}]). Call
SpecTreeValidate. Expect at least three FormatErrors:
one with rule = "name_heading", one with rule =
"dependency_targets", one with rule =
"duplicate_subsections".

#### Empty input list

Input: empty list, all_dirs = []. Call
SpecTreeValidate. Expect no format errors.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `SpecTreeValidate`.
- Use the record names from the interface:
  `SpecTreeValidateInput`, `FormatError`.
- Describe tests in terms of the functional interface —
  use function names, record names, and rule names from
  the spec.
- Each test case has: a description, setup (input data
  as list of SpecTreeValidateInput plus all_dirs list),
  actions (function call), and expected outcome.
- Input is always a list of `SpecTreeValidateInput` — no
  file I/O in tests, except for dependency_targets and
  input_target tests involving EXTERNAL/ references
  which require files on disk.
- For tests that need files on disk, describe what files
  to create and their content.
