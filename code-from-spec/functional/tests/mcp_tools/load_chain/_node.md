---
depends_on:
  - ROOT/functional/logic/mcp_tools/load_chain(interface)
output: code-from-spec/functional/tests/mcp_tools/load_chain/output.md
---

# ROOT/functional/tests/mcp_tools/load_chain

Test cases for the load chain tool.

# Public

## Test cases

All tests create a spec tree on disk with `_node.md`
files, then call `MCPLoadChain`. The result is a single
formatted string. Tests parse sections by splitting on
delimiter lines (`--- context ---`, `--- input ---`,
`--- existing artifact ---`). The first line is always
`chain_hash: <hash>`.

### Happy path

#### Simple leaf node — context and hash

Create a spec tree: SPEC (with public section containing
a `## Context` subsection with one line of content) and
SPEC/a (leaf with output, public section with a
`## Interface` subsection, agent section with content).
Do not create the output file. Call MCPLoadChain with
logical_name = "SPEC/a".

Expect result starts with `chain_hash: ` followed by a
27-character string. After `--- context ---`, the
context contains SPEC's `## subsection` headings and
their content (no `# Public` heading), the reduced
frontmatter (output only, between `---` delimiters),
SPEC/a's `## subsection` headings and their content
(no `# Public` heading), and SPEC/a's `# Agent`
heading and agent content. No `--- input ---` section.
No `--- existing artifact ---` section.

#### Ancestor public content included

Create a spec tree: SPEC (with public section), SPEC/a
(with public section), SPEC/a/b (leaf with output).
Call MCPLoadChain with "SPEC/a/b".

Expect context contains SPEC's `## subsection` headings
and their content, followed by SPEC/a's `## subsection`
headings and their content. No `# Public` headings
appear.

#### Ancestor without public section skipped

Create a spec tree: SPEC (no public section, only name
section) and SPEC/a (leaf with output and public
section). Call MCPLoadChain with "SPEC/a".

Expect context does not contain SPEC's content — it
was skipped because it has no public section.

#### Ancestor with empty public section skipped

Create a spec tree: SPEC (public section present but
empty — no content, no subsections) and SPEC/a (leaf
with output and public section). Call MCPLoadChain
with "SPEC/a".

Expect context does not contain SPEC's content.

#### Dependency without qualifier — public included

Create a spec tree: SPEC, SPEC/a (leaf with output,
depends_on = ["SPEC/b"]), SPEC/b (with public section
containing Interface and Constraints subsections). Call
MCPLoadChain with "SPEC/a".

Expect context contains SPEC/b's public content
including both subsections and their `## ` headings.

#### Dependency with qualifier — subsection only

Create a spec tree: SPEC, SPEC/a (leaf with output,
depends_on = ["SPEC/b(interface)"]), SPEC/b (with
public section containing Interface and Constraints
subsections). Call MCPLoadChain with "SPEC/a".

Expect context contains the `## Interface` heading and
its content from SPEC/b, but not the `## Constraints`
heading or its content.

#### ARTIFACT dependency — artifact tag line removed

Create a spec tree: SPEC, SPEC/a (leaf with output,
depends_on = ["ARTIFACT/b"]), SPEC/b (with
output = "out/b.go"). Create "out/b.go" with an
artifact tag line
(`// code-from-spec: SPEC/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa`)
and body content. Call MCPLoadChain with "SPEC/a".

Expect context contains the body of "out/b.go" without
the artifact tag line.

#### EXTERNAL dependency — full content

Create a spec tree: SPEC, SPEC/a (leaf with output,
depends_on = ["EXTERNAL/data/config.yaml"]). Create
"data/config.yaml" with known content. Call
MCPLoadChain with "SPEC/a".

Expect context contains the full content of
"data/config.yaml".

#### Target has reduced frontmatter with output only

Create a spec tree: SPEC, SPEC/a (leaf with depends_on
and output). Call MCPLoadChain with "SPEC/a".

Expect context contains a frontmatter block between
`---` delimiters with only the `output` field. The
`depends_on` field is not present.

#### Target agent section included

Create a spec tree: SPEC, SPEC/a (leaf with output,
public section, agent section with content). Call
MCPLoadChain with "SPEC/a".

Expect context contains SPEC/a's `## subsection`
headings and their content (no `# Public` heading),
and SPEC/a's `# Agent` heading and agent content.

#### Target without agent section — skipped

Create a spec tree: SPEC, SPEC/a (leaf with output,
public section, no agent section). Call MCPLoadChain
with "SPEC/a".

Expect no error. Context contains only public content.

#### Input present — in separate section

Create a spec tree: SPEC, SPEC/a (leaf with output,
input = "ARTIFACT/b"), SPEC/b (with output =
"out/data.json"). Create "out/data.json" with an
artifact tag line and body content. Call MCPLoadChain
with "SPEC/a".

Expect result contains `--- input ---` section with
the body of "out/data.json" without the artifact tag
line. The input content does not appear in the context
section.

#### EXTERNAL input — full content in input section

Create a spec tree: SPEC, SPEC/a (leaf with output,
input = "EXTERNAL/docs/vendor/spec.yaml"). Create
"docs/vendor/spec.yaml" with known content. Call
MCPLoadChain with "SPEC/a".

Expect result contains `--- input ---` section with
the full content of "docs/vendor/spec.yaml".

#### No input — section absent

Create a spec tree: SPEC, SPEC/a (leaf with output,
no input field). Call MCPLoadChain with "SPEC/a".

Expect result does not contain `--- input ---`.

#### Existing artifact present — in separate section

Create a spec tree: SPEC, SPEC/a (leaf with
output = "out/a.go"). Create "out/a.go" with known
content. Call MCPLoadChain with "SPEC/a".

Expect result contains `--- existing artifact ---`
section with the full content of "out/a.go".

#### Existing artifact absent — section omitted

Create a spec tree: SPEC, SPEC/a (leaf with
output = "out/a.go"). Do not create "out/a.go". Call
MCPLoadChain with "SPEC/a".

Expect result does not contain
`--- existing artifact ---`.

#### Hash is deterministic

Create a spec tree with known content. Call MCPLoadChain
twice. Expect both results have identical chain_hash.

### Error cases

#### Invalid logical name — not SPEC/

Call MCPLoadChain with "INVALID/something". Expect
error logicalnames.UnsupportedReference (propagated
from LogicalNameToPath).

#### Nonexistent node file

Call MCPLoadChain with "SPEC/nonexistent" (no _node.md
on disk). Expect error propagated from
FrontmatterParse (FileUnreadable).

#### No output declared

Create a spec tree: SPEC, SPEC/a (leaf without
output). Call MCPLoadChain with "SPEC/a". Expect
error NoOutput.

#### Invalid output path — traversal

Create a spec tree: SPEC, SPEC/a (leaf with output =
"../../etc/passwd"). Call MCPLoadChain with "SPEC/a".
Expect error InvalidOutputPath.

#### Unresolvable dependency

Create a spec tree: SPEC, SPEC/a (leaf with output,
depends_on = ["SPEC/missing"]). Do not create
SPEC/missing. Call MCPLoadChain with "SPEC/a". Expect
an error — the missing node will be detected when the
chain is processed (during hash computation or context
building).

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `MCPLoadChain`.
- The function returns a single string. Parse sections
  by splitting on delimiter lines.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `MCPLoadChain`.
- Describe setup as files to create with their
  frontmatter content.
- Use formal error names (PascalCase) as defined in the
  interface.
- When creating `_node.md` files with `# Public`
  content, all content must be under `##` subsections.
  Never place content directly under `# Public`
  without a subsection heading — this is a format
  error.
