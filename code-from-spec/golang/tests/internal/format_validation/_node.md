---
depends_on:
  - ROOT/golang/implementation/internal/node_discovery
input: ARTIFACT/golang/implementation/internal/format_validation/code(formatvalidation)
outputs:
  - id: formatvalidation_test
    path: internal/formatvalidation/formatvalidation_test.go
---

# ROOT/golang/tests/internal/format_validation

Test cases for the formatvalidation package.

# Agent

## Context

Each test uses `t.TempDir()` to create an isolated temporary
directory. Create `_node.md` files with controlled content
to simulate a spec tree. Use table-driven tests where
appropriate.

## Happy Path

### Valid leaf node passes all checks

Create a leaf node with correct heading, valid frontmatter,
and valid output paths. Expect empty `[]FormatError`.

### Valid intermediate node passes all checks

Create a parent node and a child node. The parent has only
a heading and `# Public` section (no frontmatter fields,
no `# Agent`). Expect empty `[]FormatError`.

## Failure Cases

### Heading mismatch

Create a node whose first heading does not match its
logical name. Expect a `FormatError` with rule indicating
name verification failure.

### Intermediate node with outputs

Create a parent node that has `outputs:` in frontmatter,
and a child node. Expect a `FormatError` for frontmatter
field restriction violation.

### Intermediate node with Agent section

Create a parent node that has a `# Agent` section, and a
child node. Expect a `FormatError` for agent section
restriction violation.

### depends_on targets non-existent node

Create a leaf node with `depends_on` pointing to a
non-existent logical name. Expect a `FormatError` for
dependency target validation.

### depends_on targets ancestor

Create ROOT, ROOT/a, ROOT/a/b where ROOT/a/b depends_on
ROOT. Expect a `FormatError` for redundant ancestor
dependency.

### depends_on targets descendant

Create ROOT/a with depends_on ROOT/a/b, and ROOT/a/b
exists. Expect a `FormatError` for circular descendant
dependency.

### Output path with traversal

Create a node with an output path containing `..`.
Expect a `FormatError` for output path validation.

### Duplicate public subsections

Create a node with two `## Interface` subsections under
`# Public`. Expect a `FormatError` for duplicate
subsection heading.

### Collects multiple errors

Create a node with several violations. Expect all
violations are reported, not just the first one.
