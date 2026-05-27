---
depends_on:
  - ROOT/functional/logic/mcp_tools/load_chain(interface)
outputs:
  - id: load_chain_tests
    path: code-from-spec/functional/tests/mcp_tools/load_chain/output.md
---

# ROOT/functional/tests/mcp_tools/load_chain

Test cases for the load chain tool.

# Public

## Test cases

### Happy path

#### Valid leaf node

Create a spec tree: ROOT and ROOT/a (leaf with outputs and
no dependencies). Both have public sections. Call
HandleLoadChain with logical name = "ROOT/a".

Expect success. Chain content contains ROOT with only the
body of its public section (the public heading itself is
not present), and ROOT/a with reduced frontmatter and full
body.

#### Node with dependency, no qualifier

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
referencing ROOT/b), ROOT/b (with public section containing
Interface and Constraints subsections). Call HandleLoadChain
with logical name = "ROOT/a".

Expect success. The dependency ROOT/b section contains only
its public content (both subsections).

#### Node with dependency, with qualifier

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
referencing ROOT/b(interface)), ROOT/b (with public section
containing Interface and Constraints subsections). Call
HandleLoadChain with logical name = "ROOT/a".

Expect success. The dependency ROOT/b section contains only
the Interface subsection content, not Constraints.

#### Ancestors expose only public body, without heading

Create a spec tree: ROOT (with public and private
sections), ROOT/a, ROOT/a/b (leaf with outputs). Call
HandleLoadChain with logical name = "ROOT/a/b".

Expect the sections for ROOT and ROOT/a contain only the
body of their public sections. The public heading itself,
private sections, and node name sections are not present.

#### Target has reduced frontmatter

Create a spec tree: ROOT and ROOT/a (leaf with depends_on
and outputs). Call HandleLoadChain with logical name =
"ROOT/a".

Expect the target section contains frontmatter with only
outputs. The depends_on field is not present.

#### Existing code files included in output

Create a spec tree: ROOT and ROOT/a (leaf with outputs
pointing to src/a.go). Create src/a.go with known content.
Call HandleLoadChain with logical name = "ROOT/a".

Expect success. Chain content contains a file section for
src/a.go with the file content matching what was written to
disk.

#### Non-existing code files omitted from output

Create a spec tree: ROOT and ROOT/a (leaf with outputs
pointing to src/a.go). Do not create src/a.go. Call
HandleLoadChain with logical name = "ROOT/a".

Expect success. Chain content does not contain a file
section for src/a.go.

#### Ancestor with no public section omitted

Create a spec tree: ROOT (with no public section -- only
node name and private sections) and ROOT/a (leaf with
outputs). Call HandleLoadChain with logical name = "ROOT/a".

Expect success. The chain content does not contain a file
section for ROOT.

#### Ancestor with empty public section omitted

Create a spec tree: ROOT (with a public section that has no
content and no subsections) and ROOT/a (leaf with outputs).
Call HandleLoadChain with logical name = "ROOT/a".

Expect success. The chain content does not contain a file
section for ROOT.

#### Dependency with empty extracted content omitted

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
referencing ROOT/b(interface)), ROOT/b (with a public
section containing an Interface subsection with no body).
Call HandleLoadChain with logical name = "ROOT/a".

Expect success. The chain content does not contain a file
section for ROOT/b.

#### Multiple qualifiers on same dependency consolidated

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
referencing both ROOT/b(interface) and
ROOT/b(constraints)), ROOT/b (with a public section
containing Interface and Constraints subsections, each with
distinct content). Call HandleLoadChain with logical name =
"ROOT/a".

Expect success. The chain content contains exactly one file
section for ROOT/b, and that section includes the content
of both Interface and Constraints in order, without
duplicating the file block.

### Failure cases

#### Invalid prefix

Call HandleLoadChain with logical name =
"INVALID/something". Expect error containing "target must
be a ROOT/ logical name".

#### Nonexistent spec file

Call HandleLoadChain with logical name =
"ROOT/nonexistent". Do not create the corresponding spec
file. Expect error (file not found).

#### No outputs

Create a spec tree: ROOT and ROOT/a (leaf without outputs).
Call HandleLoadChain with logical name = "ROOT/a". Expect
error containing "has no outputs".

#### Invalid outputs path -- traversal

Create a spec tree: ROOT and ROOT/a (leaf with outputs
pointing to "../../etc/passwd"). Call HandleLoadChain with
logical name = "ROOT/a". Expect error from path validation.

#### Unresolvable dependency

Create a spec tree: ROOT and ROOT/a (leaf with depends_on
referencing ROOT/b). Do not create ROOT/b's file. Call
HandleLoadChain with logical name = "ROOT/a". Expect error
from chain resolution.

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
