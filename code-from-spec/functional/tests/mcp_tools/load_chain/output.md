<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@WZZYrYkNEhF7FTv06KANQhdxUDw -->

# Test Specification: MCPLoadChain

## Test cases

All tests create a spec tree on disk with `_node.md` files, then call `MCPLoadChain`.
Results are returned as an `MCPLoadChainResult` record with fields: chain_hash, context, input.

---

### Happy path

#### Simple leaf node — context and hash

Setup:
- Create ROOT/_node.md with a public section containing one line of content.
- Create ROOT/a/_node.md with an output field, a public section with content, and an agent section with content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.chain_hash is a 27-character string.
- result.context contains ROOT's `# Public` heading and its public content.
- result.context contains a frontmatter block between `---` delimiters with only the `output` field.
- result.context contains ROOT/a's `# Public` heading and its public content.
- result.context contains ROOT/a's `# Agent` heading and its agent content.
- result.input is absent.

---

#### Ancestor public content included

Setup:
- Create ROOT/_node.md with a public section containing content.
- Create ROOT/a/_node.md with a public section containing content.
- Create ROOT/a/b/_node.md as a leaf with an output field.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a/b".

Expected outcome:
- result.context contains ROOT's `# Public` heading and public content followed by ROOT/a's `# Public` heading and public content.

---

#### Ancestor without public section skipped

Setup:
- Create ROOT/_node.md with only a name section (no public section).
- Create ROOT/a/_node.md as a leaf with an output field and a public section.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context does not contain ROOT's content — ROOT was skipped because it has no public section.

---

#### Ancestor with empty public section skipped

Setup:
- Create ROOT/_node.md with a public section that is present but contains no content and no subsections.
- Create ROOT/a/_node.md as a leaf with an output field and a public section.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context does not contain ROOT's content.

---

#### Dependency without qualifier — public included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and depends_on = ["ROOT/b"].
- Create ROOT/b/_node.md with a public section containing Interface and Constraints subsections.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains ROOT/b's public content including both the `## Interface` and `## Constraints` headings and their content.

---

#### Dependency with qualifier — subsection only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and depends_on = ["ROOT/b(interface)"].
- Create ROOT/b/_node.md with a public section containing Interface and Constraints subsections.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains the `## Interface` heading and its content from ROOT/b.
- result.context does not contain the `## Constraints` heading or its content.

---

#### ARTIFACT dependency — content minus frontmatter

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and depends_on = ["ARTIFACT/b"].
- Create ROOT/b/_node.md with output = "out/b.go".
- Create "out/b.go" with frontmatter and known body content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains the body of "out/b.go" without its frontmatter.

---

#### External file — full content

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and external = [{path: "data/config.yaml"}].
- Create "data/config.yaml" with known content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains the full content of "data/config.yaml".

---

#### Target has reduced frontmatter with output only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with a depends_on field and an output field.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains a frontmatter block between `---` delimiters with only the `output` field.
- The `depends_on` field is not present in that frontmatter block.

---

#### Target agent section included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field, a public section with content, and an agent section with content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.context contains ROOT/a's `# Public` heading and public content.
- result.context contains ROOT/a's `# Agent` heading and agent content.

---

#### Target without agent section — skipped

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and a public section, but no agent section.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- No error is raised.
- result.context contains only public content (no agent heading appears).

---

#### Input separated from context

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and input = "ARTIFACT/b".
- Create ROOT/b/_node.md with output = "out/data.json".
- Create "out/data.json" with frontmatter and known body content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.input contains the body of "out/data.json" without its frontmatter.
- The input content does not appear in result.context.

---

#### No input — field absent

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and no input field.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- result.input is absent.

---

#### Hash is deterministic

Setup:
- Create ROOT/_node.md with known content.
- Create ROOT/a/_node.md as a leaf with an output field and known content.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a" — capture result.chain_hash.
- Call MCPLoadChain again with the same logical_name — capture result.chain_hash.

Expected outcome:
- Both chain_hash values are identical.

---

### Error cases

#### Invalid logical name — not ROOT/

Setup:
- None.

Action:
- Call MCPLoadChain with logical_name = "INVALID/something".

Expected outcome:
- Error UnsupportedReference is raised (propagated from LogicalNames via LogicalNameToPath).

---

#### Nonexistent node file

Setup:
- No _node.md file exists at ROOT/nonexistent/.

Action:
- Call MCPLoadChain with logical_name = "ROOT/nonexistent".

Expected outcome:
- Error FileUnreadable is raised (propagated from FrontmatterParse via FileReader).

---

#### No output declared

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with no output field.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error NoOutput is raised.

---

#### Invalid output path — traversal

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with output = "../../etc/passwd".

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error InvalidOutputPath is raised.

---

#### Unresolvable dependency

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md as a leaf with an output field and depends_on = ["ROOT/missing"].
- Do not create ROOT/missing/_node.md.

Action:
- Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- An error is raised — the missing node is detected during chain processing (hash computation or context building).
