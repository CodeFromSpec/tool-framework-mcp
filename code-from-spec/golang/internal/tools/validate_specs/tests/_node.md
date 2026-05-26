---
depends_on:
  - ROOT/golang/internal/frontmatter
  - ROOT/golang/internal/node_discovery
input: ARTIFACT/golang/internal/tools/validate_specs/code(validate_specs)
outputs:
  - id: validate_specs_test
    path: internal/validate_specs/validate_specs_test.go
---

# ROOT/golang/internal/tools/validate_specs/tests

Tests for the validate_specs tool handler.

# Agent

## Context

Each test uses `t.TempDir()` as the project root. A spec tree
is created with the necessary `_node.md` files and frontmatter.
The working directory is changed to the temp dir for the
duration of the test so that node discovery and path validation
resolve correctly.

## Happy Path

### Clean tree with no errors

Create a spec tree with `ROOT` and `ROOT/a` (leaf with
`outputs` and valid frontmatter). Create the corresponding
output file with a matching artifact tag hash.

Expect: success result. Report contains no format errors,
no circular references, and no staleness entries.

### Detects stale artifact

Create a spec tree with `ROOT` and `ROOT/a` (leaf with
`outputs`). Create the output file with an outdated hash
in its artifact tag.

Expect: success result. Report contains a staleness entry
for `ROOT/a` with status `stale`.

### Detects missing artifact

Create a spec tree with `ROOT` and `ROOT/a` (leaf with
`outputs`). Do not create the output file.

Expect: success result. Report contains a staleness entry
for `ROOT/a` with status `missing`.

### Detects format errors

Create a spec tree with `ROOT` and `ROOT/a` where `ROOT/a`
has a malformed frontmatter or invalid `depends_on` target
that does not resolve.

Expect: success result. Report contains at least one format
error for `ROOT/a`.

### Detects circular references

Create a spec tree with `ROOT`, `ROOT/a` (depends on
`ROOT/b`), and `ROOT/b` (depends on `ROOT/a`).

Expect: success result. Report contains circular reference
entries listing the cycle participants.

### Multiple errors collected together

Create a spec tree with format errors, circular references,
and stale artifacts all present simultaneously.

Expect: success result. Report contains entries from all
three categories in a single response.

## Failure Cases

### Continues after unreadable file

Create a spec tree where one `_node.md` file has invalid
content. Other nodes are valid.

Expect: success result. The unreadable node produces a
format error. Other nodes are still validated and their
staleness is checked.
