<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@t0Hpmw8y8ZK5m2wFROBN9lSGar4 -->

## Test cases for MCPLoadChain

---

### Happy path

#### Simple leaf node — context and hash

Setup:
- Create ROOT/_node.md with a `# Public` section containing one line of content.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"`, a `# Public` section with content, and a `# Agent` section with content.
- Do not create the output file.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result starts with `chain_hash: ` followed by a 27-character string.
- After `--- context ---`, the context contains:
  - ROOT's `# Public` heading and public content.
  - ROOT/a's reduced frontmatter (only `output` field, between `---` delimiters).
  - ROOT/a's `# Public` heading and public content.
  - ROOT/a's `# Agent` heading and agent content.
- Result does not contain `--- input ---`.
- Result does not contain `--- existing artifact ---`.

---

#### Ancestor public content included

Setup:
- Create ROOT/_node.md with a `# Public` section with content.
- Create ROOT/a/_node.md with a `# Public` section with content.
- Create ROOT/a/b/_node.md with frontmatter `output: "out/b.go"`.

Action: Call MCPLoadChain with logical_name = "ROOT/a/b".

Expected outcome:
- Context contains ROOT's `# Public` heading and public content.
- Context contains ROOT/a's `# Public` heading and public content.

---

#### Ancestor without public section skipped

Setup:
- Create ROOT/_node.md with only a name section (no `# Public` section).
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and a `# Public` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context does not contain any content from ROOT (it was skipped because it has no public section).

---

#### Ancestor with empty public section skipped

Setup:
- Create ROOT/_node.md with a `# Public` section that is present but empty (no content, no subsections).
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and a `# Public` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context does not contain any content from ROOT (it was skipped because its public section is empty).

---

#### Dependency without qualifier — public included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section that has `## Interface` and `## Constraints` subsections with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains ROOT/b's public content including both `## Interface` and `## Constraints` headings and their content.

---

#### Dependency with qualifier — subsection only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `depends_on: ["ROOT/b(interface)"]`.
- Create ROOT/b/_node.md with a `# Public` section that has `## Interface` and `## Constraints` subsections with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains the `## Interface` heading and its content from ROOT/b.
- Context does not contain the `## Constraints` heading or its content.

---

#### ARTIFACT dependency — content minus frontmatter

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `depends_on: ["ARTIFACT/b"]`.
- Create ROOT/b/_node.md with frontmatter `output: "out/b.go"`.
- Create "out/b.go" with frontmatter and body content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains the body of "out/b.go" without frontmatter.

---

#### External file — full content

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `external: [{path: "data/config.yaml"}]`.
- Create "data/config.yaml" with known content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains the full content of "data/config.yaml".

---

#### Target has reduced frontmatter with output only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains a frontmatter block between `---` delimiters.
- The frontmatter block contains only the `output` field.
- The `depends_on` field is not present in the frontmatter block.

---

#### Target agent section included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"`, a `# Public` section with content, and a `# Agent` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Context contains ROOT/a's `# Public` heading and public content.
- Context contains ROOT/a's `# Agent` heading and agent content.

---

#### Target without agent section — skipped

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and a `# Public` section with content. No `# Agent` section.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- No error.
- Context contains only ROOT/a's public content (no agent heading or content).

---

#### Input present — in separate section

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `input: "ARTIFACT/b"`.
- Create ROOT/b/_node.md with frontmatter `output: "out/data.json"`.
- Create "out/data.json" with frontmatter and body content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result contains `--- input ---` section with the body of "out/data.json" without frontmatter.
- The input content does not appear in the context section.

---

#### No input — section absent

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"`. No `input` field.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result does not contain `--- input ---`.

---

#### Existing artifact present — in separate section

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"`.
- Create "out/a.go" with known content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result contains `--- existing artifact ---` section with the full content of "out/a.go".

---

#### Existing artifact absent — section omitted

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"`.
- Do not create "out/a.go".

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result does not contain `--- existing artifact ---`.

---

#### Hash is deterministic

Setup:
- Create a spec tree with known, fixed content.

Action: Call MCPLoadChain twice with the same logical_name.

Expected outcome:
- Both results have identical chain_hash values.

---

### Error cases

#### Invalid logical name — not ROOT/

Setup: None.

Action: Call MCPLoadChain with logical_name = "INVALID/something".

Expected outcome:
- Error UnsupportedReference (propagated from LogicalNames via LogicalNameToPath).

---

#### Nonexistent node file

Setup: No _node.md file on disk for the target.

Action: Call MCPLoadChain with logical_name = "ROOT/nonexistent".

Expected outcome:
- Error propagated from FrontmatterParse (FileUnreadable).

---

#### No output declared

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with no `output` field in frontmatter.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error NoOutput.

---

#### Invalid output path — traversal

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "../../etc/passwd"`.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error InvalidOutputPath.

---

#### Unresolvable dependency

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "out/a.go"` and `depends_on: ["ROOT/missing"]`.
- Do not create ROOT/missing/_node.md.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- An error is returned when the missing node is detected during chain processing (hash computation or context building).
