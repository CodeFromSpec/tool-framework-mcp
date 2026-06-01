---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate(interface)
  - ARTIFACT/golang/interfaces/os/file_reader(interface)
  - ARTIFACT/golang/interfaces/os/path_utils(interface)
  - ARTIFACT/golang/interfaces/utils/text_normalization(interface)
  - ARTIFACT/golang/interfaces/parsing/frontmatter(interface)
  - ARTIFACT/golang/interfaces/parsing/node_parsing(interface)
input: ARTIFACT/functional/tests/spec_tree/validate(format_validation_tests)
outputs:
  - id: spectreevalidate_test
    path: internal/spectreevalidate/spectreevalidate_test.go
---

# ROOT/golang/tests/spec_tree/validate

# Agent

## Test setup guidance

`SpecTreeValidate` receives `SpecTreeValidateInput`
entries built by the caller. Each entry has:
- `LogicalName`: a string like `"ROOT/a"`.
- `Frontmatter`: a `*frontmatter.Frontmatter` struct.
- `Node`: a `*parsenode.Node` struct.

The `Node` must be constructed to match what `NodeParse`
would produce:
- `NameSection.Heading`: normalized form of the logical
  name (e.g. `"root/a"` for `"ROOT/a"`). Use
  `textnormalization.NormalizeText` or hardcode the
  lowercase form.
- `NameSection.RawHeading`: the original heading line
  (e.g. `"# ROOT/a"`).
- `NameSection.Content`: `[]string` (can be empty).
- `Public`: if present, a `*parsenode.NodeSection` with
  `Heading: "public"`, `RawHeading: "# Public"`,
  `Content: []string{...}`, and `Subsections` as needed.
- `Agent`: if present, similar structure with
  `Heading: "agent"`.
- Subsection headings are also normalized (e.g.
  `"interface"` not `"Interface"`).

For tests that validate external files (fragment hashes),
use `testChdir` and create files on disk.
