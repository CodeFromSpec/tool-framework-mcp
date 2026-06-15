<!-- code-from-spec: SPEC/functional/tests/spec_tree/validate@yW9q5YAZv9MkWvDzZmLkOVsSKUI -->

# Test Specification: SpecTreeValidate

Each test case lists a description, setup (list of SpecTreeValidateInput records plus all_dirs), the action taken, and the expected outcome.

---

## Happy path

### Valid leaf node passes all checks

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"], output: "internal/out.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/b", frontmatter: {}, node: { name_section: { heading: "spec/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result is an empty list of FormatErrors

---

### Valid intermediate node passes all checks

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: { content: [], subsections: [] }, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result is an empty list of FormatErrors

---

### Leaf with no frontmatter fields

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result is an empty list of FormatErrors

---

## name_heading

### Heading matches logical name

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "name_heading"

---

### Heading does not match logical name

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/wrong" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "name_heading" }

---

## leaf_only_fields

### Intermediate node with depends_on

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "leaf_only_fields" }

---

### Intermediate node with output

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "x.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "leaf_only_fields" }

---

### Intermediate node with input

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { input: "ARTIFACT/c" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "leaf_only_fields" }

---

### Intermediate node with multiple restricted fields

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/b"], output: "x.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly two FormatErrors, both with node = "SPEC/a" and rule = "leaf_only_fields" (one per restricted field present)

---

## leaf_only_agent

### Intermediate node with agent section

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: { content: ["Agent instructions."] } } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "leaf_only_agent" }

---

### Leaf node with agent section — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: { content: ["Agent instructions."] } } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "leaf_only_agent"

---

## dependency_targets

### depends_on targets non-existent SPEC node

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/missing"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### depends_on targets ancestor

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: { depends_on: ["SPEC"] }, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a/b", rule: "dependency_targets" }

---

### depends_on targets descendant

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/a/b"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a/b", frontmatter: {}, node: { name_section: { heading: "spec/a/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/a/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### depends_on targets self

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/a"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### depends_on with valid SPEC qualifier

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["SPEC/a(interface)"] }, node: { name_section: { heading: "spec/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "dependency_targets"

---

### depends_on with valid ARTIFACT reference

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "lib.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/b", frontmatter: { depends_on: ["ARTIFACT/a"] }, node: { name_section: { heading: "spec/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "dependency_targets"

---

### depends_on with non-existent ARTIFACT reference

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["ARTIFACT/missing"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### depends_on with valid EXTERNAL reference

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["EXTERNAL/proto/api.proto"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]
- files on disk: create "proto/api.proto" with any content

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "dependency_targets"

---

### depends_on with non-existent EXTERNAL file

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["EXTERNAL/nonexistent.txt"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]
- files on disk: do not create "nonexistent.txt"

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### depends_on with unrecognized prefix

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["UNKNOWN/something"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "dependency_targets" }

---

### Multiple invalid depends_on — one error per entry

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/missing", "SPEC/also_missing"] }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly two FormatErrors, both with node = "SPEC/a" and rule = "dependency_targets" (one per invalid entry)

---

## input_target

### Valid ARTIFACT input reference

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "a.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/b", frontmatter: { input: "ARTIFACT/a" }, node: { name_section: { heading: "spec/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "input_target"

---

### Valid EXTERNAL input reference

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { input: "EXTERNAL/docs/spec.yaml" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]
- files on disk: create "docs/spec.yaml" with any content

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "input_target"

---

### Input with unsupported prefix

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { input: "SPEC/something" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "input_target" }

---

### Input references non-existent artifact

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { input: "ARTIFACT/missing" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "input_target" }

---

### Input references non-existent EXTERNAL file

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { input: "EXTERNAL/nonexistent.txt" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]
- files on disk: do not create "nonexistent.txt"

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "input_target" }

---

## missing_node_md

### Subdirectory without _node.md

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "code-from-spec/b", rule: "missing_node_md" }

---

### _-prefixed dir under code-from-spec — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/_rules"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "missing_node_md"

---

### All subdirectories have _node.md — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/b", frontmatter: {}, node: { name_section: { heading: "spec/b" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a", "code-from-spec/b"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "missing_node_md"

---

## output_paths

### Valid output path

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "internal/x.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "output_paths"

---

### Output path with traversal

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "../../etc/passwd" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "output_paths" }

---

### Output path with backslash

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { output: "internal\\x.go" }, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "output_paths" }

---

## public_subsection_required

### Public with content before first subsection

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: ["Some loose content."], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["Types."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "public_subsection_required", detail: "content in # Public must be under a ## subsection" }

---

### Public with only blank lines before subsection — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: ["", "  ", ""], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["Types."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "public_subsection_required"

---

### Public with content and no subsections

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: ["Some content."], subsections: [] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "public_subsection_required" }

---

### Public with only subsections — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: [], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["Types."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "public_subsection_required"

---

### No public section — skip

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "public_subsection_required"

---

## duplicate_subsections

### Unique subsection headings — no error

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: [], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["Types."] }, { heading: "context", raw_heading: "## Context", content: ["Background."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "duplicate_subsections"

---

### Duplicate subsection headings

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: [], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["First."] }, { heading: "interface", raw_heading: "## Interface", content: ["Second."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly one FormatError { node: "SPEC/a", rule: "duplicate_subsections" } (for the second occurrence)

---

### Three identical subsection headings

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: { content: [], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["First."] }, { heading: "interface", raw_heading: "## Interface", content: ["Second."] }, { heading: "interface", raw_heading: "## Interface", content: ["Third."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains exactly two FormatErrors, both with node = "SPEC/a" and rule = "duplicate_subsections" (for the second and third occurrences)

---

### No public section — skip

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: {}, node: { name_section: { heading: "spec/a" }, public: absent, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains no FormatError with rule = "duplicate_subsections"

---

## Cross-cutting

### Collects multiple errors from different rules

Setup:
- entries:
  - SpecTreeValidateInput { logical_name: "SPEC", frontmatter: {}, node: { name_section: { heading: "spec" }, public: absent, agent: absent } }
  - SpecTreeValidateInput { logical_name: "SPEC/a", frontmatter: { depends_on: ["SPEC/missing"] }, node: { name_section: { heading: "spec/wrong" }, public: { content: [], subsections: [{ heading: "interface", raw_heading: "## Interface", content: ["First."] }, { heading: "interface", raw_heading: "## Interface", content: ["Second."] }] }, agent: absent } }
- all_dirs: ["code-from-spec", "code-from-spec/a"]

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result contains at least three FormatErrors:
- one with node = "SPEC/a" and rule = "name_heading"
- one with node = "SPEC/a" and rule = "dependency_targets"
- one with node = "SPEC/a" and rule = "duplicate_subsections"

---

### Empty input list

Setup:
- entries: empty list
- all_dirs: []

Action: call SpecTreeValidate(entries, all_dirs)

Expected: result is an empty list of FormatErrors
