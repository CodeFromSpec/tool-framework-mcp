<!-- code-from-spec: ROOT/functional/tests/mcp_tools/load_chain@E9aGanVJe04NniIu7l1nLhqZj0M -->

# Test Specification: MCPLoadChain

## Happy Path

### TC-01: Simple leaf node — context and hash

Setup:
- Create SPEC/_node.md with:
  - frontmatter: (no depends_on, no output, no input)
  - `# Public` section containing `## Context` subsection with one line of content
- Create SPEC/a/_node.md with:
  - frontmatter: output = "out/a.txt"
  - `# Public` section containing `## Interface` subsection with content
  - `# Agent` section with content
- Do not create "out/a.txt"

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- First line matches `chain_hash: ` followed by exactly 27 characters
- String contains `--- context ---` delimiter
- After `--- context ---`:
  - SPEC's `## Context` heading and its content appear
  - No `# Public` heading appears
  - A frontmatter block appears between `---` delimiters containing only `output = "out/a.txt"`; no `depends_on` field
  - SPEC/a's `## Interface` heading and its content appear
  - No `# Public` heading appears
  - `# Agent` heading and SPEC/a's agent content appear
- String does not contain `--- input ---`
- String does not contain `--- existing artifact ---`

---

### TC-02: Ancestor public content included

Setup:
- Create SPEC/_node.md with `# Public` section and `## Overview` subsection
- Create SPEC/a/_node.md with `# Public` section and `## Description` subsection
- Create SPEC/a/b/_node.md with frontmatter: output = "out/b.txt", and `# Public` section

Actions:
- Call MCPLoadChain("SPEC/a/b")

Expected outcome:
- After `--- context ---`:
  - SPEC's `## Overview` heading and its content appear
  - SPEC/a's `## Description` heading and its content appear

---

### TC-03: Ancestor without public section skipped

Setup:
- Create SPEC/_node.md with only a `# Name` section (no `# Public` section)
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", `# Public` section with `## Summary` subsection

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, SPEC's content does not appear
- SPEC/a's `## Summary` heading and content appear

---

### TC-04: Ancestor with empty public section skipped

Setup:
- Create SPEC/_node.md with a `# Public` section that has no subsections and no content
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", `# Public` section with `## Summary` subsection

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, SPEC's content does not appear
- SPEC/a's `## Summary` heading and content appear

---

### TC-05: Dependency without qualifier — full public included

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["SPEC/b"]
- Create SPEC/b/_node.md with `# Public` section containing `## Interface` and `## Constraints` subsections with content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, SPEC/b's `## Interface` heading and content appear
- SPEC/b's `## Constraints` heading and content appear

---

### TC-06: Dependency with qualifier — subsection only

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["SPEC/b(interface)"]
- Create SPEC/b/_node.md with `# Public` section containing `## Interface` and `## Constraints` subsections with content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, SPEC/b's `## Interface` heading and content appear
- SPEC/b's `## Constraints` heading and content do not appear

---

### TC-07: ARTIFACT dependency — artifact tag line removed

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["ARTIFACT/b"]
- Create SPEC/b/_node.md with frontmatter: output = "out/b.go"
- Create "out/b.go" with:
  - First line: `// code-from-spec: SPEC/b@aaaaaaaaaaaaaaaaaaaaaaaaaaa`
  - Remaining lines: body content (known text)

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, the body content of "out/b.go" appears
- The artifact tag line does not appear in the context

---

### TC-08: EXTERNAL dependency — full content

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["EXTERNAL/data/config.yaml"]
- Create "data/config.yaml" with known content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, the full content of "data/config.yaml" appears

---

### TC-09: Target has reduced frontmatter with output only

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["SPEC/b"]
- Create SPEC/b/_node.md with `# Public` section with `## Info` subsection

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, a frontmatter block between `---` delimiters contains only the `output` field
- The `depends_on` field does not appear in the frontmatter block

---

### TC-10: Target agent section included

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with:
  - frontmatter: output = "out/a.txt"
  - `# Public` section with `## Interface` subsection
  - `# Agent` section with content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- After `--- context ---`, SPEC/a's `## Interface` heading and content appear
- `# Agent` heading and the agent content appear
- No `# Public` heading appears

---

### TC-11: Target without agent section — skipped

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with:
  - frontmatter: output = "out/a.txt"
  - `# Public` section with `## Interface` subsection
  - No `# Agent` section

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- No error
- Context contains SPEC/a's public content
- No `# Agent` heading appears in the result

---

### TC-12: Input present — in separate section

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", input = "ARTIFACT/b"
- Create SPEC/b/_node.md with frontmatter: output = "out/data.json"
- Create "out/data.json" with:
  - First line: artifact tag line
  - Remaining lines: body content (known text)

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Result contains `--- input ---` delimiter
- After `--- input ---`, the body content of "out/data.json" appears
- The artifact tag line does not appear
- The input content does not appear in the section before `--- input ---`

---

### TC-13: EXTERNAL input — full content in input section

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", input = "EXTERNAL/docs/vendor/spec.yaml"
- Create "docs/vendor/spec.yaml" with known content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Result contains `--- input ---` delimiter
- After `--- input ---`, the full content of "docs/vendor/spec.yaml" appears

---

### TC-14: No input — section absent

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", no input field

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Result does not contain `--- input ---`

---

### TC-15: Existing artifact present — in separate section

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.go"
- Create "out/a.go" with known content

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Result contains `--- existing artifact ---` delimiter
- After `--- existing artifact ---`, the full content of "out/a.go" appears

---

### TC-16: Existing artifact absent — section omitted

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.go"
- Do not create "out/a.go"

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Result does not contain `--- existing artifact ---`

---

### TC-17: Hash is deterministic

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", `# Public` section with `## Notes` subsection with fixed content

Actions:
- Call MCPLoadChain("SPEC/a") — call once, record chain_hash
- Call MCPLoadChain("SPEC/a") — call again, record chain_hash

Expected outcome:
- Both chain_hash values are identical

---

## Error Cases

### TC-18: Invalid logical name — not SPEC/

Actions:
- Call MCPLoadChain("INVALID/something")

Expected outcome:
- Error logicalnames.UnsupportedReference is raised

---

### TC-19: Nonexistent node file

Setup:
- Do not create any _node.md at SPEC/nonexistent

Actions:
- Call MCPLoadChain("SPEC/nonexistent")

Expected outcome:
- Error propagated from FrontmatterParse (FileUnreadable) is raised

---

### TC-20: No output declared

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter containing no output field

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Error NoOutput is raised

---

### TC-21: Invalid output path — traversal

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "../../etc/passwd"

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- Error InvalidOutputPath is raised

---

### TC-22: Unresolvable dependency

Setup:
- Create SPEC/_node.md with no public section
- Create SPEC/a/_node.md with frontmatter: output = "out/a.txt", depends_on = ["SPEC/missing"]
- Do not create SPEC/missing/_node.md

Actions:
- Call MCPLoadChain("SPEC/a")

Expected outcome:
- An error is raised when the missing node is encountered during chain processing (hash computation or context building)
