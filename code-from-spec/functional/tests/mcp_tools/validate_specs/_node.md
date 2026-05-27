---
depends_on:
  - ROOT/functional/logic/mcp_tools/validate_specs(interface)
outputs:
  - id: validate_specs_tests
    path: code-from-spec/functional/tests/mcp_tools/validate_specs/output.md
---

# ROOT/functional/tests/mcp_tools/validate_specs

Test cases for the validate specs tool.

# Public

## Test cases

### Happy path

#### Clean tree with no errors

Create a spec tree with ROOT and ROOT/a (leaf with outputs
and valid frontmatter). Create the corresponding output
file with a matching artifact tag hash. Call
HandleValidateSpecs.

Expect success. Report contains no format errors, no
circular references, and no staleness entries.

#### Detects stale artifact

Create a spec tree with ROOT and ROOT/a (leaf with
outputs). Create the output file with an outdated hash in
its artifact tag. Call HandleValidateSpecs.

Expect success. Report contains a staleness entry for
ROOT/a with status "stale".

#### Detects missing artifact

Create a spec tree with ROOT and ROOT/a (leaf with
outputs). Do not create the output file. Call
HandleValidateSpecs.

Expect success. Report contains a staleness entry for
ROOT/a with status "missing".

#### Detects format errors

Create a spec tree with ROOT and ROOT/a where ROOT/a has
invalid frontmatter or an unresolvable depends_on target.
Call HandleValidateSpecs.

Expect success. Report contains at least one format error
for ROOT/a.

#### Detects circular references

Create a spec tree with ROOT, ROOT/a (depends on ROOT/b),
and ROOT/b (depends on ROOT/a). Call HandleValidateSpecs.

Expect success. Report contains circular reference entries
listing the cycle participants.

#### Multiple errors collected together

Create a spec tree with format errors, circular references,
and stale artifacts all present simultaneously. Call
HandleValidateSpecs.

Expect success. Report contains entries from all three
categories in a single response.

### Failure cases

#### Continues after unreadable file

Create a spec tree where one node file has invalid content.
Other nodes are valid. Call HandleValidateSpecs.

Expect success. The unreadable node produces a format
error. Other nodes are still validated and their staleness
is checked.

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
