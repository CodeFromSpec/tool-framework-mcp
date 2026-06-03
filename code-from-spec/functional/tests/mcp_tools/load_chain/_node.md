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

Create a spec tree: ROOT (with public section containing
one line of content) and ROOT/a (leaf with output,
public section with content, agent section with content).
Do not create the output file. Call MCPLoadChain with
logical_name = "ROOT/a".

Expect result starts with `chain_hash: ` followed by a
27-character string. After `--- context ---`, the
context contains ROOT's `# Public` heading and public
content, the reduced frontmatter (output only, between
`---` delimiters), ROOT/a's `# Public` heading and
public content, and ROOT/a's `# Agent` heading and
agent content. No `--- input ---` section. No
`--- existing artifact ---` section.

#### Ancestor public content included

Create a spec tree: ROOT (with public section), ROOT/a
(with public section), ROOT/a/b (leaf with output).
Call MCPLoadChain with "ROOT/a/b".

Expect context contains ROOT's `# Public` heading and
public content followed by ROOT/a's `# Public` heading
and public content.

#### Ancestor without public section skipped

Create a spec tree: ROOT (no public section, only name
section) and ROOT/a (leaf with output and public
section). Call MCPLoadChain with "ROOT/a".

Expect context does not contain ROOT's content — it
was skipped because it has no public section.

#### Ancestor with empty public section skipped

Create a spec tree: ROOT (public section present but
empty — no content, no subsections) and ROOT/a (leaf
with output and public section). Call MCPLoadChain
with "ROOT/a".

Expect context does not contain ROOT's content.

#### Dependency without qualifier — public included

Create a spec tree: ROOT, ROOT/a (leaf with output,
depends_on = ["ROOT/b"]), ROOT/b (with public section
containing Interface and Constraints subsections). Call
MCPLoadChain with "ROOT/a".

Expect context contains ROOT/b's public content
including both subsections and their `## ` headings.

#### Dependency with qualifier — subsection only

Create a spec tree: ROOT, ROOT/a (leaf with output,
depends_on = ["ROOT/b(interface)"]), ROOT/b (with
public section containing Interface and Constraints
subsections). Call MCPLoadChain with "ROOT/a".

Expect context contains the `## Interface` heading and
its content from ROOT/b, but not the `## Constraints`
heading or its content.

#### ARTIFACT dependency — content minus frontmatter

Create a spec tree: ROOT, ROOT/a (leaf with output,
depends_on = ["ARTIFACT/b"]), ROOT/b (with
output = "out/b.go"). Create "out/b.go" with
frontmatter and body content. Call MCPLoadChain with
"ROOT/a".

Expect context contains the body of "out/b.go" without
frontmatter.

#### External file — full content

Create a spec tree: ROOT, ROOT/a (leaf with output,
external = [{path: "data/config.yaml"}]). Create
"data/config.yaml" with known content. Call
MCPLoadChain with "ROOT/a".

Expect context contains the full content of
"data/config.yaml".

#### Target has reduced frontmatter with output only

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
and output). Call MCPLoadChain with "ROOT/a".

Expect context contains a frontmatter block between
`---` delimiters with only the `output` field. The
`depends_on` field is not present.

#### Target agent section included

Create a spec tree: ROOT, ROOT/a (leaf with output,
public section, agent section with content). Call
MCPLoadChain with "ROOT/a".

Expect context contains ROOT/a's `# Public` heading
and public content, and ROOT/a's `# Agent` heading
and agent content.

#### Target without agent section — skipped

Create a spec tree: ROOT, ROOT/a (leaf with output,
public section, no agent section). Call MCPLoadChain
with "ROOT/a".

Expect no error. Context contains only public content.

#### Input present — in separate section

Create a spec tree: ROOT, ROOT/a (leaf with output,
input = "ARTIFACT/b"), ROOT/b (with output =
"out/data.json"). Create "out/data.json" with
frontmatter and body. Call MCPLoadChain with "ROOT/a".

Expect result contains `--- input ---` section with
the body of "out/data.json" without frontmatter. The
input content does not appear in the context section.

#### No input — section absent

Create a spec tree: ROOT, ROOT/a (leaf with output,
no input field). Call MCPLoadChain with "ROOT/a".

Expect result does not contain `--- input ---`.

#### Existing artifact present — in separate section

Create a spec tree: ROOT, ROOT/a (leaf with
output = "out/a.go"). Create "out/a.go" with known
content. Call MCPLoadChain with "ROOT/a".

Expect result contains `--- existing artifact ---`
section with the full content of "out/a.go".

#### Existing artifact absent — section omitted

Create a spec tree: ROOT, ROOT/a (leaf with
output = "out/a.go"). Do not create "out/a.go". Call
MCPLoadChain with "ROOT/a".

Expect result does not contain
`--- existing artifact ---`.

#### Hash is deterministic

Create a spec tree with known content. Call MCPLoadChain
twice. Expect both results have identical chain_hash.

### Error cases

#### Invalid logical name — not ROOT/

Call MCPLoadChain with "INVALID/something". Expect
error UnsupportedReference (propagated from
LogicalNames via LogicalNameToPath).

#### Nonexistent node file

Call MCPLoadChain with "ROOT/nonexistent" (no _node.md
on disk). Expect error propagated from
FrontmatterParse (FileUnreadable).

#### No output declared

Create a spec tree: ROOT, ROOT/a (leaf without
output). Call MCPLoadChain with "ROOT/a". Expect
error NoOutput.

#### Invalid output path — traversal

Create a spec tree: ROOT, ROOT/a (leaf with output =
"../../etc/passwd"). Call MCPLoadChain with "ROOT/a".
Expect error InvalidOutputPath.

#### Unresolvable dependency

Create a spec tree: ROOT, ROOT/a (leaf with output,
depends_on = ["ROOT/missing"]). Do not create
ROOT/missing. Call MCPLoadChain with "ROOT/a". Expect
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
