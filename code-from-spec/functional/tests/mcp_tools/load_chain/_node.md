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

All tests create a spec tree on disk with `_node.md`
files, then call `MCPLoadChain`.

### Happy path

#### Simple leaf node — context and hash

Create a spec tree: ROOT (with public section containing
one line of content) and ROOT/a (leaf with outputs,
public section with content, agent section with content).
Call MCPLoadChain with logical_name = "ROOT/a".

Expect result has:
- `chain_hash`: a 27-character string
- `context`: contains ROOT's public content (without
  the `# Public` heading), the reduced frontmatter
  (outputs only, between `---` delimiters), ROOT/a's
  public content, and ROOT/a's agent content
- `input`: absent

#### Ancestor public content included

Create a spec tree: ROOT (with public section), ROOT/a
(with public section), ROOT/a/b (leaf with outputs).
Call MCPLoadChain with "ROOT/a/b".

Expect context contains ROOT's public content followed
by ROOT/a's public content, both without `# Public`
heading.

#### Ancestor without public section skipped

Create a spec tree: ROOT (no public section, only name
section) and ROOT/a (leaf with outputs and public
section). Call MCPLoadChain with "ROOT/a".

Expect context does not contain ROOT's content — it
was skipped because it has no public section.

#### Ancestor with empty public section skipped

Create a spec tree: ROOT (public section present but
empty — no content, no subsections) and ROOT/a (leaf
with outputs and public section). Call MCPLoadChain
with "ROOT/a".

Expect context does not contain ROOT's content.

#### Dependency without qualifier — public included

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
depends_on = ["ROOT/b"]), ROOT/b (with public section
containing Interface and Constraints subsections). Call
MCPLoadChain with "ROOT/a".

Expect context contains ROOT/b's public content
including both subsections and their `## ` headings.

#### Dependency with qualifier — subsection only

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
depends_on = ["ROOT/b(interface)"]), ROOT/b (with
public section containing Interface and Constraints
subsections). Call MCPLoadChain with "ROOT/a".

Expect context contains only the Interface subsection
content from ROOT/b, not Constraints.

#### ARTIFACT dependency — content minus frontmatter

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
depends_on = ["ARTIFACT/b(code)"]), ROOT/b (with
outputs = [{id: "code", path: "out/b.go"}]). Create
"out/b.go" with frontmatter and body content. Call
MCPLoadChain with "ROOT/a".

Expect context contains the body of "out/b.go" without
frontmatter.

#### External file — full content

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
external = [{path: "data/config.yaml"}]). Create
"data/config.yaml" with known content. Call
MCPLoadChain with "ROOT/a".

Expect context contains the full content of
"data/config.yaml".

#### External file with fragments — line ranges only

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
external = [{path: "data/big.txt", fragments:
[{lines: "2-4", hash: "ignored"}]}]). Create
"data/big.txt" with 10 lines. Call MCPLoadChain with
"ROOT/a".

Expect context contains only lines 2-4 from
"data/big.txt".

#### Target has reduced frontmatter with outputs only

Create a spec tree: ROOT, ROOT/a (leaf with depends_on
and outputs). Call MCPLoadChain with "ROOT/a".

Expect context contains a frontmatter block between
`---` delimiters with only the `outputs` field. The
`depends_on` field is not present.

#### Target agent section included

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
public section, agent section with content). Call
MCPLoadChain with "ROOT/a".

Expect context contains both the public and agent
content of ROOT/a (without `# Public` and `# Agent`
headings).

#### Target without agent section — skipped

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
public section, no agent section). Call MCPLoadChain
with "ROOT/a".

Expect no error. Context contains only public content.

#### Input separated from context

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
input = "ARTIFACT/b(data)"), ROOT/b (with outputs =
[{id: "data", path: "out/data.json"}]). Create
"out/data.json" with frontmatter and body. Call
MCPLoadChain with "ROOT/a".

Expect result.input contains the body of "out/data.json"
without frontmatter. The input content does not appear
in result.context.

#### No input — field absent

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
no input field). Call MCPLoadChain with "ROOT/a".

Expect result.input is absent.

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

#### No outputs declared

Create a spec tree: ROOT, ROOT/a (leaf without
outputs). Call MCPLoadChain with "ROOT/a". Expect
error NoOutputs.

#### Invalid output path — traversal

Create a spec tree: ROOT, ROOT/a (leaf with outputs
pointing to "../../etc/passwd"). Call MCPLoadChain
with "ROOT/a". Expect error InvalidOutputPath.

#### Unresolvable dependency

Create a spec tree: ROOT, ROOT/a (leaf with outputs,
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
- Use the record name from the interface:
  `MCPLoadChainResult`.
- Each test case creates a spec tree on disk with
  `_node.md` files, then calls `MCPLoadChain`.
- Describe setup as files to create with their
  frontmatter content.
- Use formal error names (PascalCase) as defined in the
  interface.
