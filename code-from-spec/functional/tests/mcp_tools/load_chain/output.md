<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@6zMrgGQ7JQF7MjwKF_hDRoFxbtM -->

## Happy path

### Simple leaf node — context and hash

Setup:
- Create ROOT/_node.md with a `# Public` section containing one line of content.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go`, a `# Public` section with content, and a `# Agent` section with content.
- Do not create "out/a.go".

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Result starts with `chain_hash: ` followed by exactly 27 characters.
- After `--- context ---`: context contains ROOT's `# Public` heading and its content, a reduced frontmatter block between `---` delimiters containing only `output: out/a.go`, ROOT/a's `# Public` heading and its content, ROOT/a's `# Agent` heading and its content.
- Result does not contain `--- input ---`.
- Result does not contain `--- existing artifact ---`.

---

### Ancestor public content included

Setup:
- Create ROOT/_node.md with a `# Public` section.
- Create ROOT/a/_node.md with a `# Public` section.
- Create ROOT/a/b/_node.md with frontmatter `output: out/b.go` and a `# Public` section.

Action: call MCPLoadChain with logical_name = "ROOT/a/b".

Expected:
- Context contains ROOT's `# Public` heading and content followed by ROOT/a's `# Public` heading and content.

---

### Ancestor without public section skipped

Setup:
- Create ROOT/_node.md with only a `# Name` section (no `# Public` section).
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and a `# Public` section with content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context does not contain ROOT's heading or content.

---

### Ancestor with empty public section skipped

Setup:
- Create ROOT/_node.md with a `# Public` section that has no content, no subsections.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and a `# Public` section with content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context does not contain ROOT's content.

---

### Dependency without qualifier — public included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section containing `## Interface` and `## Constraints` subsections with content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains ROOT/b's `# Public` heading, its `## Interface` heading and content, and its `## Constraints` heading and content.

---

### Dependency with qualifier — subsection only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `depends_on: ["ROOT/b(interface)"]`.
- Create ROOT/b/_node.md with a `# Public` section containing `## Interface` and `## Constraints` subsections with content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains the `## Interface` heading and its content from ROOT/b.
- Context does not contain the `## Constraints` heading or its content.

---

### ARTIFACT dependency — content minus frontmatter

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `depends_on: ["ARTIFACT/b"]`.
- Create ROOT/b/_node.md with frontmatter `output: out/b.go`.
- Create "out/b.go" with a frontmatter block followed by body content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains the body of "out/b.go" without the frontmatter block.

---

### External file — full content

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `external: [{path: "data/config.yaml"}]`.
- Create "data/config.yaml" with known content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains the full content of "data/config.yaml".

---

### Target has reduced frontmatter with output only

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `depends_on: ["ROOT/b"]`.
- Create ROOT/b/_node.md with a `# Public` section.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains a frontmatter block between `---` delimiters with only `output: out/a.go`.
- The `depends_on` field is not present in the frontmatter block.

---

### Target agent section included

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go`, a `# Public` section with content, and a `# Agent` section with content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Context contains ROOT/a's `# Public` heading and content.
- Context contains ROOT/a's `# Agent` heading and content.

---

### Target without agent section — skipped

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and a `# Public` section with content. No `# Agent` section.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- No error.
- Context contains the public content.
- Context does not contain a `# Agent` heading.

---

### Input present — in separate section

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `input: ARTIFACT/b`.
- Create ROOT/b/_node.md with frontmatter `output: out/data.json`.
- Create "out/data.json" with a frontmatter block followed by body content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Result contains `--- input ---` section.
- Content after `--- input ---` is the body of "out/data.json" without its frontmatter.
- The input content does not appear in the context section (before `--- input ---`).

---

### No input — section absent

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go`. No `input` field.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Result does not contain `--- input ---`.

---

### Existing artifact present — in separate section

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go`.
- Create "out/a.go" with known content.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Result contains `--- existing artifact ---` section.
- Content after `--- existing artifact ---` is the full content of "out/a.go".

---

### Existing artifact absent — section omitted

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go`.
- Do not create "out/a.go".

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Result does not contain `--- existing artifact ---`.

---

### Hash is deterministic

Setup:
- Create a spec tree with known, fixed content: ROOT/_node.md with a `# Public` section and ROOT/a/_node.md with frontmatter `output: out/a.go` and a `# Public` section.

Action: call MCPLoadChain with logical_name = "ROOT/a" twice.

Expected:
- Both calls return results with identical `chain_hash` values.

---

## Error cases

### Invalid logical name — not ROOT/

Setup: none.

Action: call MCPLoadChain with logical_name = "INVALID/something".

Expected:
- Error logicalnames.UnsupportedReference (propagated from LogicalNameToPath).

---

### Nonexistent node file

Setup: none (no _node.md on disk for ROOT/nonexistent).

Action: call MCPLoadChain with logical_name = "ROOT/nonexistent".

Expected:
- Error propagated from FrontmatterParse indicating the file is unreadable (FileUnreadable).

---

### No output declared

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with a `# Public` section and no `output` field in frontmatter.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Error NoOutput.

---

### Invalid output path — traversal

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: ../../etc/passwd`.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- Error InvalidOutputPath.

---

### Unresolvable dependency

Setup:
- Create ROOT/_node.md.
- Create ROOT/a/_node.md with frontmatter `output: out/a.go` and `depends_on: ["ROOT/missing"]`.
- Do not create ROOT/missing/_node.md.

Action: call MCPLoadChain with logical_name = "ROOT/a".

Expected:
- An error is returned. The missing node is detected during chain processing (hash computation or context building).
