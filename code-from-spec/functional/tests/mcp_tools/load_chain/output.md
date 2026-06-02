<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@mCGe3SyrwZEGestcEAgtmVH2WqA -->

## Happy path

### Simple leaf node — context and hash

Setup:
- Create ROOT/_node.md with a `# Public` section containing one line of content.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go`, a `# Public` section with content, and a `# Agent` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Result is an MCPLoadChainResult.
- `chain_hash` is a 27-character string.
- `context` contains, in order:
  - ROOT's `# Public` heading and its public content.
  - A frontmatter block between `---` delimiters containing only the `output` field.
  - ROOT/a's `# Public` heading and its public content.
  - ROOT/a's `# Agent` heading and its agent content.
- `input` is absent.

---

### Ancestor public content included

Setup:
- Create ROOT/_node.md with a `# Public` section with content.
- Create ROOT/a/_node.md with a `# Public` section with content.
- Create ROOT/a/b/_node.md with frontmatter `output: some/output.go`.

Action: Call MCPLoadChain with logical_name = "ROOT/a/b".

Expected outcome:
- `context` contains ROOT's `# Public` heading and public content followed by ROOT/a's `# Public` heading and public content.

---

### Ancestor without public section skipped

Setup:
- Create ROOT/_node.md with only a `# Name` section (no `# Public` section).
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and a `# Public` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` does not contain any content from ROOT's `_node.md`.

---

### Ancestor with empty public section skipped

Setup:
- Create ROOT/_node.md with a `# Public` section that is present but contains no content and no subsections.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and a `# Public` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` does not contain any content from ROOT's `_node.md`.

---

### Dependency without qualifier — public included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section containing an `## Interface` subsection and a `## Constraints` subsection, each with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains ROOT/b's `# Public` content including both the `## Interface` heading and content and the `## Constraints` heading and content.

---

### Dependency with qualifier — subsection only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `depends_on: ["ROOT/b(interface)"]`.
- Create ROOT/b/_node.md with a `# Public` section containing an `## Interface` subsection and a `## Constraints` subsection, each with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains the `## Interface` heading and its content from ROOT/b.
- `context` does not contain the `## Constraints` heading or its content.

---

### ARTIFACT dependency — content minus frontmatter

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `depends_on: ["ARTIFACT/b"]`.
- Create ROOT/b/_node.md with frontmatter `output: out/b.go`.
- Create "out/b.go" with frontmatter and body content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains the body of "out/b.go" without its frontmatter.

---

### External file — full content

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `external: [{path: "data/config.yaml"}]`.
- Create "data/config.yaml" with known content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains the full content of "data/config.yaml".

---

### Target has reduced frontmatter with output only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains a frontmatter block between `---` delimiters with only the `output` field.
- The `depends_on` field is not present in the frontmatter block.

---

### Target agent section included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go`, a `# Public` section with content, and a `# Agent` section with content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `context` contains ROOT/a's `# Public` heading and public content.
- `context` contains ROOT/a's `# Agent` heading and agent content.

---

### Target without agent section — skipped

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and a `# Public` section with content. No `# Agent` section.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- No error is returned.
- `context` contains only the public content. No `# Agent` heading is present.

---

### Input separated from context

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `input: "ARTIFACT/b"`.
- Create ROOT/b/_node.md with frontmatter `output: out/data.json`.
- Create "out/data.json" with frontmatter and body content.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `result.input` contains the body of "out/data.json" without frontmatter.
- The input content does not appear in `result.context`.

---

### No input — field absent

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go`. No `input` field.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- `result.input` is absent.

---

### Hash is deterministic

Setup:
- Create a spec tree with known content: ROOT/_node.md with a `# Public` section, ROOT/a/_node.md with frontmatter `output: some/output.go` and a `# Public` section.

Action: Call MCPLoadChain with logical_name = "ROOT/a" twice.

Expected outcome:
- Both results have identical `chain_hash` values.

---

## Error cases

### Invalid logical name — not ROOT/

Action: Call MCPLoadChain with logical_name = "INVALID/something".

Expected outcome:
- Error UnsupportedReference is returned (propagated from LogicalNames via LogicalNameToPath).

---

### Nonexistent node file

Action: Call MCPLoadChain with logical_name = "ROOT/nonexistent" where no `_node.md` exists on disk.

Expected outcome:
- Error FileUnreadable is returned (propagated from FrontmatterParse).

---

### No output declared

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with a `# Public` section but no `output` field in frontmatter.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error NoOutput is returned.

---

### Invalid output path — traversal

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: "../../etc/passwd"`.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- Error InvalidOutputPath is returned.

---

### Unresolvable dependency

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: some/output.go` and `depends_on: ["ROOT/missing"]`.
- Do not create ROOT/missing/_node.md.

Action: Call MCPLoadChain with logical_name = "ROOT/a".

Expected outcome:
- An error is returned. The missing node is detected during chain processing (hash computation or context building).
