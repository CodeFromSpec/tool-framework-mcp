---
depends_on:
  - ROOT/functional/logic/utils/format_validation(interface)
outputs:
  - id: format_validation_tests
    path: code-from-spec/functional/tests/utils/format_validation/output.md
---

# ROOT/functional/tests/utils/format_validation

Test cases for the format validation component.

# Public

## Test cases

### Happy path

#### Valid leaf node passes all checks

Create a leaf node with correct heading, valid frontmatter,
and valid output paths. Call ValidateFormat. Expect no
format errors returned.

#### Valid intermediate node passes all checks

Create a parent node and a child node. The parent has only
a heading and public section (no frontmatter fields, no
agent section). Call ValidateFormat. Expect no format errors
returned.

### Failure cases

#### Heading mismatch

Create a node whose first heading does not match its
logical name. Expect a format error with rule indicating
name verification failure.

#### Intermediate node with outputs

Create a parent node that has outputs in frontmatter, and
a child node. Expect a format error for frontmatter field
restriction violation.

#### Intermediate node with agent section

Create a parent node that has an agent section, and a child
node. Expect a format error for agent section restriction
violation.

#### depends_on targets non-existent node

Create a leaf node with depends_on pointing to a
non-existent logical name. Expect a format error for
dependency target validation.

#### depends_on targets ancestor

Create ROOT, ROOT/a, ROOT/a/b where ROOT/a/b depends_on
ROOT. Expect a format error for redundant ancestor
dependency.

#### depends_on targets descendant

Create ROOT/a with depends_on ROOT/a/b, and ROOT/a/b
exists. Expect a format error for circular descendant
dependency.

#### Output path with traversal

Create a node with an output path containing `..`. Expect
a format error for output path validation.

#### Duplicate public subsections

Create a node with two subsections of the same name under
the public section. Expect a format error for duplicate
subsection heading.

#### Collects multiple errors

Create a node with several violations. Expect all
violations are reported, not just the first one.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
